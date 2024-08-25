// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.747
package main

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"fmt"
	"github.com/danielhep/go-elections/internal/types"
	"slices"
	"time"
)

func contestPage(contest types.Contest, candidates []types.Candidate, countyUpdates []types.Update, stateUpdate *types.Update) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Var2 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
			templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
			if !templ_7745c5c3_IsBuffer {
				defer func() {
					templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
					if templ_7745c5c3_Err == nil {
						templ_7745c5c3_Err = templ_7745c5c3_BufErr
					}
				}()
			}
			ctx = templ.InitializeContext(ctx)
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"bg-white shadow overflow-hidden sm:rounded-lg mb-6\"><div class=\"px-4 py-5 sm:px-6\"><h2 class=\"text-xl leading-6 font-medium text-gray-900\">Ballot Title: ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var3 string
			templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(contest.Name)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/web/contestPage.templ`, Line: 14, Col: 88}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</h2><h3 class=\"text-lg leading-6 text-gray-700 mt-1\">District: ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var4 string
			templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(contest.District)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/web/contestPage.templ`, Line: 15, Col: 81}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</h3></div><div class=\"border-t border-gray-200 px-4 py-5 sm:p-0\"><div class=\"overflow-x-auto\"><table class=\"min-w-full divide-y divide-gray-200\"><thead class=\"bg-gray-50\"><tr><th class=\"px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider\">Date</th>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			for _, candidate := range candidates {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<th class=\"px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var5 string
				templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(candidate.Name)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/web/contestPage.templ`, Line: 24, Col: 116}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</th>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</tr></thead> <tbody class=\"bg-white divide-y divide-gray-200\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if stateUpdate != nil {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<tr><td class=\"px-6 py-4 whitespace-nowrap text-sm text-gray-900 flex gap-2\"><img src=\"/static/stateflag.jpg\" class=\"h-6 rounded-md\"><p class=\"pt-[2px]\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var6 string
				templ_7745c5c3_Var6, templ_7745c5c3_Err = templ.JoinStringErrs(formatFirstCol(*stateUpdate))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/web/contestPage.templ`, Line: 33, Col: 60}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var6))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</p></td>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				for _, candidate := range candidates {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<td class=\"px-6 py-4 whitespace-nowrap text-sm text-gray-500\">")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var7 string
					templ_7745c5c3_Var7, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprintf("%d", getVotesForUpdate(candidate, stateUpdate.ID)))
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/web/contestPage.templ`, Line: 37, Col: 76}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var7))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</td>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</tr>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			for _, update := range countyUpdates {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<tr><td class=\"px-6 py-4 whitespace-nowrap text-sm text-gray-900 flex gap-2\"><img src=\"/static/kingcounty.jpg\" class=\"h-6 rounded-md\"><p class=\"pt-[2px]\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var8 string
				templ_7745c5c3_Var8, templ_7745c5c3_Err = templ.JoinStringErrs(formatFirstCol(update))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/web/contestPage.templ`, Line: 46, Col: 54}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var8))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</p></td>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				for _, candidate := range candidates {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<td class=\"px-6 py-4 whitespace-nowrap text-sm text-gray-500\">")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					var templ_7745c5c3_Var9 string
					templ_7745c5c3_Var9, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprintf("%d", getVotesForUpdate(candidate, update.ID)))
					if templ_7745c5c3_Err != nil {
						return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/web/contestPage.templ`, Line: 50, Col: 71}
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var9))
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</td>")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</tr>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</tbody></table></div></div></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if len(countyUpdates) > 0 {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div id=\"chart-data\" class=\"bg-white shadow overflow-hidden sm:rounded-lg\" chart-data=\"")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var10 string
				templ_7745c5c3_Var10, templ_7745c5c3_Err = templ.JoinStringErrs(templ.JSONString(getChartData(candidates)))
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `cmd/web/contestPage.templ`, Line: 64, Col: 59}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var10))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" x-data=\"{data: JSON.parse(document.getElementById(&#39;chart-data&#39;).getAttribute(&#39;chart-data&#39;))}\" x-init=\"\n\t\t\t\tconst ctx = document.getElementById(&#39;voteChart&#39;).getContext(&#39;2d&#39;);\n    \t\t\tnew Chart(ctx, {\n    \t\t\t\ttype: &#39;line&#39;,\n    \t\t\t\tdata: data,\n    \t\t\t\toptions: {\n    \t\t\t\t\tresponsive: true,\n    \t\t\t\t\tplugins: {\n    \t\t\t\t\t\tlegend: {\n    \t\t\t\t\t\t\tposition: &#39;top&#39;,\n    \t\t\t\t\t\t},\n    \t\t\t\t\t\ttitle: {\n    \t\t\t\t\t\t\tdisplay: true,\n    \t\t\t\t\t\t\ttext: &#39;King County Votes Over Time&#39;\n    \t\t\t\t\t\t}\n    \t\t\t\t\t}\n    \t\t\t\t}\n    \t\t\t});\n            \"><canvas id=\"voteChart\" width=\"400\" height=\"200\"></canvas></div>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" <div class=\"mt-4\"><a href=\"/\" class=\"text-indigo-600 hover:text-indigo-900\">Back to Main Page</a></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return templ_7745c5c3_Err
		})
		templ_7745c5c3_Err = layout(contest.Name+" Results").Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

func getAllUpdatesFromCandidate(candidate types.Candidate) (ret []types.Update) {
	for _, voteTally := range candidate.VoteTallies {
		if voteTally.Votes > 0 {
			ret = append(ret, voteTally.Update)
		}
	}
	return
}

func getCandidates(voteTallies []types.VoteTally) []types.Candidate {
	candidateMap := make(map[uint]types.Candidate)
	for _, tally := range voteTallies {
		candidateMap[tally.CandidateID] = tally.Candidate
	}

	var candidates []types.Candidate
	for _, candidate := range candidateMap {
		candidates = append(candidates, candidate)
	}
	return candidates
}

func getVotesForUpdate(candidate types.Candidate, updateID uint) int {
	for _, tally := range candidate.VoteTallies {
		if tally.UpdateID == updateID {
			return tally.Votes
		}
	}
	return 0
}

func formatFirstCol(update types.Update) string {
	if update.JurisdictionType == types.CountyJurisdiction {
		return formatDate(update.Timestamp)
	} else {
		return "State (final tally)"
	}
}

func formatDate(timestamp time.Time) string {
	return timestamp.Format("Jan 02, 2006")
}

type chartData struct {
	Labels   []string `json:"labels"`
	Datasets []struct {
		Label           string `json:"label"`
		Data            []int  `json:"data"`
		BorderColor     string `json:"borderColor"`
		BackgroundColor string `json:"backgroundColor"`
		Fill            bool   `json:"fill"`
	} `json:"datasets"`
}

func getChartData(candidates []types.Candidate) chartData {
	// Prepare data for the chart
	datasets := make(map[string][]int)
	var labels []string

	for i, candidate := range candidates {
		for _, voteTally := range candidate.VoteTallies {
			if voteTally.Update.JurisdictionType == types.StateJurisdiction {
				continue
			}
			if i == 0 {
				labels = append(labels, formatDate(voteTally.Update.Timestamp))
			}
			datasets[candidate.Name] = append(datasets[candidate.Name], voteTally.Votes)
		}
	}

	labels = slices.Clip(labels)

	// Create Chart.js data structure
	chartData := chartData{
		Labels: labels,
	}

	colors := []string{"#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF", "#FF9F40"}
	colorIndex := 0

	for candidate, votes := range datasets {
		chartData.Datasets = append(chartData.Datasets, struct {
			Label           string `json:"label"`
			Data            []int  `json:"data"`
			BorderColor     string `json:"borderColor"`
			BackgroundColor string `json:"backgroundColor"`
			Fill            bool   `json:"fill"`
		}{
			Label:           candidate,
			Data:            votes,
			BorderColor:     colors[colorIndex%len(colors)],
			BackgroundColor: colors[colorIndex%len(colors)],
			Fill:            false,
		})
		colorIndex++
	}

	return chartData
}
