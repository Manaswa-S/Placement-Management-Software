package gocharts

import (
	"fmt"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	sqlc "go.mod/internal/sqlc/generate"
)

var (
	sankeyNode = []opts.SankeyNode{}
	sankeyLink = []opts.SankeyLink{}
)

func SankeyApplications(data *sqlc.ApplicationsStatusCountsRow) *charts.Sankey {

	zeroLinkval := float32(0.05)

	hc := float32(data.HiredCount)
	oc := hc + float32(data.OfferedCount)
	slc := hc + oc + float32(data.ShortlistedCount)
	urc := hc + oc + slc + float32(data.UnderReviewCount)
	ac := float32(data.AppliedCount)
	rc := float32(data.RejectedCount)
	rcURC := urc - slc - float32(data.UnderReviewCount)
	rcSLC := slc - oc - float32(data.ShortlistedCount)

	sankeyNode = []opts.SankeyNode{
		{Name: "Applied", Value: fmt.Sprintf("%f", ac), Depth: opts.Int(0)},
		// {Name: "NoUpdate", Value: fmt.Sprintf("%d", ), Depth: opts.Int(1)},
		{Name: "UnderReview", Value: fmt.Sprintf("%f", urc), Depth: opts.Int(1)},
		{Name: "Shortlisted", Value: fmt.Sprintf("%f", slc), Depth: opts.Int(2)},
		{Name: "Rejected", Value: fmt.Sprintf("%f", rc), Depth: opts.Int(3)},
		{Name: "Offered", Value: fmt.Sprintf("%f", oc), Depth: opts.Int(3)},
		{Name: "Hired", Value: fmt.Sprintf("%f", hc), Depth: opts.Int(4)},
	}

	sankeyLink = []opts.SankeyLink{
		{Source: "Applied", Target: "UnderReview", Value: float32(max(urc, zeroLinkval))},
		{Source: "UnderReview", Target: "Shortlisted", Value: float32(max(slc, zeroLinkval))},
		{Source: "UnderReview", Target: "Rejected", Value: float32(max(rcURC, zeroLinkval))},
		{Source: "Shortlisted", Target: "Offered", Value: float32(max(oc, zeroLinkval))},
		{Source: "Shortlisted", Target: "Rejected", Value: float32(max(rcSLC, zeroLinkval))},
		{Source: "Offered", Target: "Hired", Value: float32(max(hc, zeroLinkval))},
	}


	sankey := charts.NewSankey()
	
	sankey.AddSeries("Applications", sankeyNode, sankeyLink)
	sankey.SetGlobalOptions(
		// charts.WithToolboxOpts(opts.Toolbox{
		// 	Show: opts.Bool(true),
		// 	Orient: "vertical",
		// 	Feature: &opts.ToolBoxFeature{
				// DataView: &opts.ToolBoxFeatureDataView{
				// 	Show:  opts.Bool(true),        
				// 	Title: "Data View",    
				// 	Lang: []string{"Data View", "Close", "Refresh"},
				// },

				// Restore: &opts.ToolBoxFeatureRestore{
				// 	Show:  opts.Bool(true),
				// 	Title: "Restore",
				// },

				// SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
				// 	Show: opts.Bool(true),
				// 	Title: "applicationsSankey",
				// },
			// },
		// }),

		charts.WithTitleOpts(opts.Title{
			Title: "Applications",
		}),
	)

	return sankey
}
