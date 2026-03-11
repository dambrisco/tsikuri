package match

import (
	"image"
	"math"
	"runtime"
	"sort"
	"sync"

	"github.com/dambrisco/tsikuri/backend"
)

// Matcher implements backend.MatchBackend using normalized cross-correlation,
// with a fallback to normalized SSD for uniform templates.
type Matcher struct{}

// New creates a new Matcher.
func New() *Matcher {
	return &Matcher{}
}

// FindBest finds the single best match of needle in haystack.
func (m *Matcher) FindBest(haystack, needle image.Image, minScore float64) (*backend.MatchResult, error) {
	hGray := ToGray(haystack)
	nGray := ToGray(needle)

	hBounds := haystack.Bounds()
	nBounds := needle.Bounds()
	tw, th := nBounds.Dx(), nBounds.Dy()
	hw, hh := hBounds.Dx(), hBounds.Dy()

	if tw > hw || th > hh {
		return nil, nil
	}

	tMean, tStd := templateStats(nGray, tw, th)
	uniform := tStd < 1e-10
	scorer := makeScorer(hGray, nGray, tw, th, tMean, tStd, uniform)

	searchW := hw - tw + 1
	searchH := hh - th + 1

	bestScore, bestX, bestY := parallelBest(searchW, searchH, scorer)

	if bestScore < minScore {
		return nil, nil
	}

	return &backend.MatchResult{
		Score: bestScore,
		Rect: image.Rect(
			hBounds.Min.X+bestX,
			hBounds.Min.Y+bestY,
			hBounds.Min.X+bestX+tw,
			hBounds.Min.Y+bestY+th,
		),
	}, nil
}

// FindAll finds all non-overlapping matches with score >= minScore.
func (m *Matcher) FindAll(haystack, needle image.Image, minScore float64) ([]backend.MatchResult, error) {
	hGray := ToGray(haystack)
	nGray := ToGray(needle)

	hBounds := haystack.Bounds()
	nBounds := needle.Bounds()
	tw, th := nBounds.Dx(), nBounds.Dy()
	hw, hh := hBounds.Dx(), hBounds.Dy()

	if tw > hw || th > hh {
		return nil, nil
	}

	tMean, tStd := templateStats(nGray, tw, th)
	uniform := tStd < 1e-10
	scorer := makeScorer(hGray, nGray, tw, th, tMean, tStd, uniform)

	searchW := hw - tw + 1
	searchH := hh - th + 1

	all := parallelCollect(searchW, searchH, minScore, scorer)

	// Sort by score descending
	sort.Slice(all, func(i, j int) bool {
		return all[i].score > all[j].score
	})

	// Non-maximum suppression
	var results []backend.MatchResult
	used := make([]bool, len(all))

	for i, c := range all {
		if used[i] {
			continue
		}
		r := image.Rect(
			hBounds.Min.X+c.x,
			hBounds.Min.Y+c.y,
			hBounds.Min.X+c.x+tw,
			hBounds.Min.Y+c.y+th,
		)
		results = append(results, backend.MatchResult{Score: c.score, Rect: r})

		for j := i + 1; j < len(all); j++ {
			if used[j] {
				continue
			}
			other := image.Rect(
				hBounds.Min.X+all[j].x,
				hBounds.Min.Y+all[j].y,
				hBounds.Min.X+all[j].x+tw,
				hBounds.Min.Y+all[j].y+th,
			)
			overlap := r.Intersect(other)
			if !overlap.Empty() {
				overlapArea := overlap.Dx() * overlap.Dy()
				templateArea := tw * th
				if float64(overlapArea)/float64(templateArea) > 0.5 {
					used[j] = true
				}
			}
		}
	}

	return results, nil
}

// scoreFunc computes a similarity score at position (x, y).
type scoreFunc func(x, y int) float64

// makeScorer returns the appropriate scoring function based on template uniformity.
func makeScorer(hGray, nGray [][]float64, tw, th int, tMean, tStd float64, uniform bool) scoreFunc {
	if uniform {
		// For uniform templates, use normalized SSD:
		// score = 1 - sqrt(mean_squared_error) where MSE is relative to the template mean
		return func(x, y int) float64 {
			return ssdScore(hGray, tw, th, x, y, tMean)
		}
	}
	return func(x, y int) float64 {
		return nccAt(hGray, nGray, x, y, tw, th, tMean, tStd)
	}
}

type candidate struct {
	score float64
	x, y  int
}

func numWorkers(searchH int) int {
	n := runtime.NumCPU()
	if n > searchH {
		n = searchH
	}
	if n < 1 {
		n = 1
	}
	return n
}

// parallelBest finds the single best scoring position.
func parallelBest(searchW, searchH int, scorer scoreFunc) (bestScore float64, bestX, bestY int) {
	nw := numWorkers(searchH)
	results := make([]candidate, nw)
	var wg sync.WaitGroup
	rowsPerWorker := (searchH + nw - 1) / nw

	for w := 0; w < nw; w++ {
		wg.Add(1)
		startY := w * rowsPerWorker
		endY := startY + rowsPerWorker
		if endY > searchH {
			endY = searchH
		}
		go func(workerIdx, sy, ey int) {
			defer wg.Done()
			local := candidate{score: -1}
			for y := sy; y < ey; y++ {
				for x := 0; x < searchW; x++ {
					s := scorer(x, y)
					if s > local.score {
						local = candidate{score: s, x: x, y: y}
					}
				}
			}
			results[workerIdx] = local
		}(w, startY, endY)
	}
	wg.Wait()

	bestScore = -1
	for _, r := range results {
		if r.score > bestScore {
			bestScore = r.score
			bestX = r.x
			bestY = r.y
		}
	}
	return
}

// parallelCollect finds all positions with score >= minScore.
func parallelCollect(searchW, searchH int, minScore float64, scorer scoreFunc) []candidate {
	nw := numWorkers(searchH)
	workerResults := make([][]candidate, nw)
	var wg sync.WaitGroup
	rowsPerWorker := (searchH + nw - 1) / nw

	for w := 0; w < nw; w++ {
		wg.Add(1)
		startY := w * rowsPerWorker
		endY := startY + rowsPerWorker
		if endY > searchH {
			endY = searchH
		}
		go func(workerIdx, sy, ey int) {
			defer wg.Done()
			var local []candidate
			for y := sy; y < ey; y++ {
				for x := 0; x < searchW; x++ {
					s := scorer(x, y)
					if s >= minScore {
						local = append(local, candidate{score: s, x: x, y: y})
					}
				}
			}
			workerResults[workerIdx] = local
		}(w, startY, endY)
	}
	wg.Wait()

	var all []candidate
	for _, wr := range workerResults {
		all = append(all, wr...)
	}
	return all
}

// nccAt computes the NCC score at position (ox, oy) in the haystack.
func nccAt(haystack, needle [][]float64, ox, oy, tw, th int, tMean, tStd float64) float64 {
	var hSum float64
	for y := 0; y < th; y++ {
		for x := 0; x < tw; x++ {
			hSum += haystack[oy+y][ox+x]
		}
	}
	n := float64(tw * th)
	hMean := hSum / n

	var cc, hVar float64
	for y := 0; y < th; y++ {
		for x := 0; x < tw; x++ {
			hd := haystack[oy+y][ox+x] - hMean
			td := needle[y][x] - tMean
			cc += hd * td
			hVar += hd * hd
		}
	}

	hStd := math.Sqrt(hVar / n)
	if hStd < 1e-10 {
		// Haystack patch is also uniform — check if they match
		if math.Abs(hMean-tMean) < 0.01 {
			return 1.0
		}
		return 0
	}

	return cc / (n * hStd * tStd)
}

// ssdScore computes a similarity score for a uniform template using
// normalized sum of squared differences. Returns 1.0 for a perfect match.
func ssdScore(haystack [][]float64, tw, th, ox, oy int, tMean float64) float64 {
	n := float64(tw * th)
	var ssd float64
	for y := 0; y < th; y++ {
		for x := 0; x < tw; x++ {
			d := haystack[oy+y][ox+x] - tMean
			ssd += d * d
		}
	}
	// Normalize: max possible SSD per pixel is 1.0 (when comparing 0 and 1)
	mse := ssd / n
	// Convert to similarity: 1.0 = perfect, 0.0 = worst
	return 1.0 - math.Sqrt(mse)
}

// templateStats computes the mean and standard deviation of the template.
func templateStats(tmpl [][]float64, w, h int) (mean, std float64) {
	n := float64(w * h)
	var sum float64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sum += tmpl[y][x]
		}
	}
	mean = sum / n

	var variance float64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			d := tmpl[y][x] - mean
			variance += d * d
		}
	}
	std = math.Sqrt(variance / n)
	return
}
