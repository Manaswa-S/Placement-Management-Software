package gocharts

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"go.mod/internal/dto"
)

var (
	marksFreq []opts.BarData
)

// ResultDraft creates the actual charts for the test cumulative result.
// Most of the params are hard coded except the data values.
func ResultDraft(data *dto.CumulativeChartsData) (*components.Page, error) {
	// create a new chart page
	page := components.NewPage()
	
	// create charts 
	marksbar, err := marksBar(data.Xaxis, data.Yaxis)
	if err != nil {
		return nil, err
	}

	passfailPie, err := passfailPie(data.PassCount, data.FailCount) 
	if err != nil {
		return nil, err
	}

	// add them to the page
	page.AddCharts(marksbar, passfailPie)
	
	return page, nil
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

func passfailPie(pass int64, fail int64) (*charts.Pie, error) {

	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Pass Fail Count"}),
		charts.WithColorsOpts(opts.Colors{"green", "red"}),
	)

	pie.AddSeries("pie", []opts.PieData{
		{Name: "Pass", Value: pass},
		{Name: "Fail", Value: fail},
	}).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{
			Show: opts.Bool(true),
			Formatter: "{b} : {c}",
		}),
		charts.WithPieChartOpts(opts.PieChart{
			Radius: []string{"45%", "75%"},
		}),
	)


	return pie, nil
}