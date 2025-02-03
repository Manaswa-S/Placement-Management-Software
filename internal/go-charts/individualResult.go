package gocharts

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"go.mod/internal/dto"
)


func IndividualResult(data *dto.IndividualChartsData) (*components.Page, error) {

	qcountFunnel, err := qCountFunnel(data.FunnelDimensions, data.FunnelValues)
	if err != nil {
		return nil, err
	}

	accuracyRadar, err := accuracyRadar(data.RadarNames, data.RadarValues)
	if err != nil {
		return nil, err
	}


	page := components.NewPage()
	page.AddCharts(qcountFunnel, accuracyRadar)

	return page, nil
}

func qCountFunnel(dimensions []string, values []int64) (*charts.Funnel, error) {

	qFunnel := charts.NewFunnel()
	qFunnel.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Total VS Attempted VS Correct",}),
	)

	funnelData := make([]opts.FunnelData, 0)
	for i, d := range dimensions {
		funnelData = append(funnelData, opts.FunnelData{Name: d, Value: values[i]})
	}

	qFunnel.AddSeries("Insights", funnelData)

	return qFunnel, nil
}

func accuracyRadar(indicators []*opts.Indicator, values []float32) (*charts.Radar, error) {

	radar := charts.NewRadar()

	radarData := []opts.RadarData{{Value: values}}

	radar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Radar",}),
		charts.WithRadarComponentOpts(opts.RadarComponent{
			Indicator: indicators,
			SplitArea: &opts.SplitArea{Show: opts.Bool(true)},
			SplitLine: &opts.SplitLine{Show: opts.Bool(true)},
		}),
	)

	radar.AddSeries("Insights", radarData)

	return radar, nil
}