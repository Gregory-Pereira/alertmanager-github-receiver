// Copyright 2017 alertmanager-github-receiver Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//////////////////////////////////////////////////////////////////////////////

package alerts

import (
	"fmt"
	"html/template"
	"strings"
	"testing"

	"github.com/prometheus/alertmanager/notify/webhook"
	amtmpl "github.com/prometheus/alertmanager/template"
)

func checkForBadCharacters(basestr string, badchars ...string) (bool) {
	for _, badchar := range badchars {
		if strings.Contains(basestr, badchar) {
			return true
		}
	}
	return false
}

func TestFormatIssueBody(t *testing.T) {
	wh := createWebhookMessage("FakeAlertName", "firing", "")
	brokenTemplate := `
		{{range .NOT_REAL_FIELD}}
		    * {{.Status}}
		{{end}}
	`
	alertTemplate = template.Must(template.New("alert").Parse(brokenTemplate))
	got := formatIssueBody(wh)
	if got != "" {
		t.Errorf("formatIssueBody() = %q, want empty string", got)
	}
}

func TestFormatTitleSimple(t *testing.T) {
	msg := webhook.Message{
		Data: &amtmpl.Data{
			Status: "firing",
			Alerts: []amtmpl.Alert{
				{
					Annotations: amtmpl.KV{"env": "prod", "svc": "foo"},
				},
				{
					Annotations: amtmpl.KV{"env": "stage", "svc": "foo", "testing": "true"},
				},
			},
		},
	}
	tests := []struct {
		testName	 string
		titleTmpl    string
		expectErrTxt string
		expectOutput string
	}{
		{
			testName: "Success-issue-title-one",
			titleTmpl: `{{ .Data.Status }}`,
			expectErrTxt: "",
			expectOutput: "firing",
		},
		{
			testName: "Success-issue-title-two",
			titleTmpl: `Success-testing-issue-2`,
			expectErrTxt: "",
			expectOutput: "Success-testing-issue-2",
		},
		{
			testName: "Failure-issue-title-one-parsing-bad-chars",
			titleTmpl: `Failure-issue-title-one\"@#$%\n`,
			expectErrTxt: `parsing error.`,
			expectOutput: ``,
		},
		{
			testName: "Success-issue-title-four",
			titleTmpl: ``,
			expectErrTxt: "",
			expectOutput: "Issue created",
		},
		{
			testName: "Unknown-case",
			titleTmpl: `{{ .Data.Potato }}`,
			expectErrTxt: "can't evaluate field Potato",
			expectOutput: "",
		},
	}
	string_interpolation_variable := "issue variable name"
	tests[3].titleTmpl = fmt.Sprintf("Success-issue-title-with-string-interpolation: %s", string_interpolation_variable)

	for testNum, tc := range tests {
		testName := fmt.Sprintf("tc=%d", testNum)
		t.Run(testName, func(t *testing.T) {
			var extraLabels []string
			var labelTemplatelList []string
			var githubRepo = "default"
			var autoClose = true
			var resolvedLabel string
			rh, err := NewReceiver(&fakeClient{}, githubRepo, autoClose, resolvedLabel, extraLabels, tc.titleTmpl, labelTemplatelList)
			if err != nil {
				t.Fatal(err)
			}

			title, err := rh.formatTitle(&msg)
			if tc.expectErrTxt == "" && err != nil {
				t.Error(err)
			}
			if tc.expectErrTxt != "" {
				if err == nil {
					if checkForBadCharacters(title, "\n", "\\", "\"") {
						t.Error("parsing error.")
					} else {
						t.Error()
					}
				} else if !strings.Contains(err.Error(), tc.expectErrTxt) {
					t.Error(err.Error())
				}
			}
			if tc.expectOutput == "" && title != "" && tc.{
				fmt.Print("ERROR Case 3: ", err)
				t.Error(title)
			}
			fmt.Print("Test Number:", testNum, "\nTitle:", title, "\nExpectedOutput:", tc.expectOutput, "\nExpectedErr:", tc.expectErrTxt, "\nBooleanContains:", strings.Contains(title, tc.expectOutput), "\nTEST END------------------\n\n")
			if !strings.Contains(title, tc.expectOutput) {
				fmt.Print("ERROR Case 4: ", err)
				t.Error(title)
			}
		})
	}
}

func TestFormatLabels(t *testing.T) {
	msg := webhook.Message{
		Data: &amtmpl.Data{
			Status: "firing",
			Alerts: []amtmpl.Alert{
				{
					Labels: amtmpl.KV{"env": "prod", "foo": "bar", "fooTwo": "barTwo"},
				},
				{
					Labels: amtmpl.KV{"env": "staging", "cluster": "rick", "namespace": "openshift-monitoring", "application": "grafana"},
				},
			},
		},
	}
	tests := []struct {
		testName 	 string
		labelsTmpl   string
		expectErrTxt string
		expectOutput string
	}{
		{
			testName: "Success-case-one",
			labelsTmpl: ` template: | {{ range .Alerts }} {{.Labels.foo}} {{- end}}`,
			expectErrTxt: "",
			expectOutput: "bar",
		},
		{
			testName: "success-case-two",
			labelsTmpl: ` template: | {{ .Data.Status }}`,
			expectErrTxt: "",
			expectOutput: "firing",
		},
		{
			testName: "Success-no-severity-label-exists",
			labelsTmpl: ` template: | {{if .Labels.severity}} {{.Labels.severity}} {{- end}}`,
			expectErrTxt: "", 
			expectOutput: "",
		},
		{
			testName: "Failure-cant-evaluate-field-severity", 
			labelsTmpl: ` template: | {{.Labels.severity}}`,
			expectErrTxt: "can't evaluate field severity",
			expectOutput: "",
		},
	}
	for testNum, tc := range tests {
		testName := fmt.Sprintf("tc=%d", testNum)
		t.Run(testName, func(t *testing.T) {
			var extraLabels []string
			var githubRepo = "default"
			var autoClose = true
			var resolvedLabel string
			var testLabelsTemplate []string
			testLabelsTemplate[0] = tc.labelsTmpl
			rh, err := NewReceiver(&fakeClient{}, githubRepo, autoClose, resolvedLabel, extraLabels, tc.testName, testLabelsTemplate)
			if err != nil {
				t.Fatal(err)
			}
			labels, err := rh.formatLabels(&msg)
			if tc.expectErrTxt == "" && err != nil {
				t.Error(err)
			}
			if tc.expectErrTxt != "" {
				if err == nil {
					t.Error()
				} else if !strings.Contains(err.Error(), tc.expectErrTxt) {
					t.Error(err.Error())
				}
			}
			if tc.expectOutput == "" && len(labels) != 0 {
				t.Error(rh.TitleTmpl)
			}
			for _, label := range labels {
				if !strings.Contains(label, tc.expectOutput) {
					t.Error(rh.TitleTmpl)
				}
			}
		})
	}
}


