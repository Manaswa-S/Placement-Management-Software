package gocharts

import (
	"io"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

var (
	marksFreq []opts.BarData
)

func CumulativeResultCharts(xaxis []string, yaxis []int64) (string, error) {

	page := components.NewPage()

	f, err := os.Create("./template/common/result/cumulative.html")
	if err != nil {
		return "", err
	}
	defer f.Close()


	marksbar, err := marksBar(xaxis, yaxis)
	if err != nil {
		return "", err
	}
	_ = marksbar

	err = page.Render(io.MultiWriter(f))
	if err != nil {
		return "", err
	}

	return "", nil
}


func marksBar(xaxis []string, yaxis []int64) (*charts.Bar, error) {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "BAR GRAPH",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Width: "80%",
			Height: "500px",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Marks",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "No. of Students",
		}),

	)

	for _, y := range yaxis {
		marksFreq = append(marksFreq, opts.BarData{Value: y})
	}

	bar.SetXAxis(xaxis).AddSeries("Marks", marksFreq)

	

	return bar, nil
}