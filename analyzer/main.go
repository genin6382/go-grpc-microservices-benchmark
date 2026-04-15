package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"math"
	"os"
	"sort"
	"strconv"
)

type Row struct {
	Timestamp string
	Strategy  string
	Pattern   string
	Iter      int
	Path      string
	UserID    string
	Status    int
	LatencyMS float64
	Error     string
}

type Summary struct {
	Strategy     string  `json:"strategy"`
	Pattern      string  `json:"pattern"`
	Requests     int     `json:"requests"`
	SuccessRate  float64 `json:"success_rate"`
	AvgLatencyMS float64 `json:"avg_latency_ms"`
	P50LatencyMS float64 `json:"p50_latency_ms"`
	P95LatencyMS float64 `json:"p95_latency_ms"`
	P99LatencyMS float64 `json:"p99_latency_ms"`
	ErrorCount   int     `json:"error_count"`
}

type DashboardData struct {
	Summaries []Summary
	JSON      template.JS
}

func main() {
	in := flag.String("in", "output/results.csv", "input csv")
	summaryOut := flag.String("summary", "output/summary.csv", "summary csv")
	dashboardOut := flag.String("dashboard", "output/dashboard.html", "dashboard html")
	flag.Parse()

	rows, err := readRows(*in)
	if err != nil {
		panic(err)
	}

	summaries := summarize(rows)

	if err := writeSummaryCSV(*summaryOut, summaries); err != nil {
		panic(err)
	}
	if err := writeDashboard(*dashboardOut, summaries); err != nil {
		panic(err)
	}
}

func readRows(path string) ([]Row, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var rows []Row
	for i, rec := range records {
		if i == 0 {
			continue
		}
		iter, _ := strconv.Atoi(rec[3])
		status, _ := strconv.Atoi(rec[6])
		latency, _ := strconv.ParseFloat(rec[7], 64)

		rows = append(rows, Row{
			Timestamp: rec[0],
			Strategy:  rec[1],
			Pattern:   rec[2],
			Iter:      iter,
			Path:      rec[4],
			UserID:    rec[5],
			Status:    status,
			LatencyMS: latency,
			Error:     rec[8],
		})
	}
	return rows, nil
}

func summarize(rows []Row) []Summary {
	grouped := map[string][]Row{}
	for _, row := range rows {
		key := row.Strategy + "|" + row.Pattern
		grouped[key] = append(grouped[key], row)
	}

	summaries := make([]Summary, 0, len(grouped))
	for _, group := range grouped {
		latencies := make([]float64, 0, len(group))
		success := 0
		errors := 0
		total := 0.0

		for _, row := range group {
			latencies = append(latencies, row.LatencyMS)
			total += row.LatencyMS

			if row.Error == "" && row.Status >= 200 && row.Status < 500 {
				success++
			}
			if row.Error != "" || row.Status >= 500 {
				errors++
			}
		}

		sort.Float64s(latencies)

		count := len(group)
		avg := 0.0
		if count > 0 {
			avg = total / float64(count)
		}

		successRate := 0.0
		if count > 0 {
			successRate = float64(success) * 100 / float64(count)
		}

		summaries = append(summaries, Summary{
			Strategy:     group[0].Strategy,
			Pattern:      group[0].Pattern,
			Requests:     count,
			SuccessRate:  round2(successRate),
			AvgLatencyMS: round2(avg),
			P50LatencyMS: round2(percentile(latencies, 50)),
			P95LatencyMS: round2(percentile(latencies, 95)),
			P99LatencyMS: round2(percentile(latencies, 99)),
			ErrorCount:   errors,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Strategy == summaries[j].Strategy {
			return summaries[i].Pattern < summaries[j].Pattern
		}
		return summaries[i].Strategy < summaries[j].Strategy
	})

	return summaries
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}

	rank := (p / 100.0) * float64(len(sorted)-1)
	lo := int(math.Floor(rank))
	hi := int(math.Ceil(rank))
	if lo == hi {
		return sorted[lo]
	}

	weight := rank - float64(lo)
	return sorted[lo]*(1-weight) + sorted[hi]*weight
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func writeSummaryCSV(path string, summaries []Summary) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	_ = w.Write([]string{
		"strategy", "pattern", "requests", "success_rate",
		"avg_latency_ms", "p50_latency_ms", "p95_latency_ms", "p99_latency_ms", "error_count",
	})

	for _, s := range summaries {
		_ = w.Write([]string{
			s.Strategy,
			s.Pattern,
			fmt.Sprint(s.Requests),
			fmt.Sprintf("%.2f", s.SuccessRate),
			fmt.Sprintf("%.2f", s.AvgLatencyMS),
			fmt.Sprintf("%.2f", s.P50LatencyMS),
			fmt.Sprintf("%.2f", s.P95LatencyMS),
			fmt.Sprintf("%.2f", s.P99LatencyMS),
			fmt.Sprint(s.ErrorCount),
		})
	}
	return nil
}

func writeDashboard(path string, summaries []Summary) error {
	raw, err := json.Marshal(summaries)
	if err != nil {
		return err
	}

	data := DashboardData{
		Summaries: summaries,
		JSON:      template.JS(raw),
	}

	const tpl = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Benchmark Dashboard</title>
  <script src="https://cdn.plot.ly/plotly-2.35.2.min.js"></script>
  <style>
    body { font-family: Arial, sans-serif; margin: 24px; background: #f6f8fb; color: #222; }
    h1 { margin-bottom: 12px; }
    .card { background: white; border-radius: 12px; padding: 16px; box-shadow: 0 4px 18px rgba(0,0,0,.08); margin-bottom: 20px; }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(320px, 1fr)); gap: 20px; }
    table { width: 100%; border-collapse: collapse; }
    th, td { padding: 10px 12px; border-bottom: 1px solid #eee; text-align: left; }
    th { background: #eef2f7; }
  </style>
</head>
<body>
  <h1>Load Balancer Benchmark Dashboard</h1>

  <div class="card">
    <table>
      <thead>
        <tr>
          <th>Strategy</th>
          <th>Pattern</th>
          <th>Requests</th>
          <th>Success %</th>
          <th>Avg ms</th>
          <th>p50 ms</th>
          <th>p95 ms</th>
          <th>p99 ms</th>
          <th>Errors</th>
        </tr>
      </thead>
      <tbody>
        {{range .Summaries}}
        <tr>
          <td>{{.Strategy}}</td>
          <td>{{.Pattern}}</td>
          <td>{{.Requests}}</td>
          <td>{{printf "%.2f" .SuccessRate}}</td>
          <td>{{printf "%.2f" .AvgLatencyMS}}</td>
          <td>{{printf "%.2f" .P50LatencyMS}}</td>
          <td>{{printf "%.2f" .P95LatencyMS}}</td>
          <td>{{printf "%.2f" .P99LatencyMS}}</td>
          <td>{{.ErrorCount}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>

  <div class="grid">
    <div class="card"><div id="p95"></div></div>
    <div class="card"><div id="success"></div></div>
    <div class="card"><div id="avg"></div></div>
  </div>

<script>
  const data = {{.JSON}};
  const labels = data.map(d => d.strategy + " | " + d.pattern);

  Plotly.newPlot("p95", [{
    x: labels,
    y: data.map(d => d.p95_latency_ms),
    type: "bar",
    marker: { color: "#2563eb" }
  }], { title: "p95 Latency (ms)" });

  Plotly.newPlot("success", [{
    x: labels,
    y: data.map(d => d.success_rate),
    type: "bar",
    marker: { color: "#16a34a" }
  }], { title: "Success Rate (%)", yaxis: { range: [0, 100] } });

  Plotly.newPlot("avg", [{
    x: labels,
    y: data.map(d => d.avg_latency_ms),
    type: "bar",
    marker: { color: "#ea580c" }
  }], { title: "Average Latency (ms)" });
</script>
</body>
</html>`

	t, err := template.New("dashboard").Parse(tpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, data)
}