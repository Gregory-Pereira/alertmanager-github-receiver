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
					Annotations: amtmpl.KV{"env": "stage", "svc": "foo"},
				},
			},
		},
	}
	tests := []struct {
		testName	 string
		tmplTxt      string
		expectErrTxt string
		expectOutput string
	}{
		{
			testName: "Success-check-annotation-foo",
			tmplTxt: "foo",
			expectErrTxt: "",
			expectOutput: "foo",
		},
		{
			testName: "Success-test-data-status-firing",
			tmplTxt: "{{ .Data.Status }}",
			expectErrTxt: "",
			expectOutput: "firing",
		},
		{
			testName: "Succes-test-status-firing",
			tmplTxt: "{{ .Status }}",
			expectErrTxt: "",
			expectOutput: "firing",
		},
		{
			testName: "Success-test-environment",
			tmplTxt: "{{ range .Alerts }}{{ .Annotations.env }} {{ end }}",
			expectErrTxt: "",
			expectOutput: "prod stage ",
		},
		{
			testName: "Failure-test-improper-label-name",
			tmplTxt: "{{ .Foo }}",
			expectErrTxt: "can't evaluate field Foo",
			expectOutput: "",
		},
	}

	for testNum, tc := range tests {
		testName := fmt.Sprintf("tc=%d", testNum)
		t.Run(testName, func(t *testing.T) {
			var extraLabels []string
			var labelTmplList []string
			var githubRepo = "default"
			var autoClose = true
			var resolvedLabel string
			rh, err := NewReceiver(&fakeClient{}, githubRepo, autoClose, resolvedLabel, extraLabels, titleTmpl, labelTmplList)
			if err != nil {
				t.Fatal(err)
			}

			title, err := rh.formatTitle(&msg)
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
			if tc.expectOutput == "" && title != "" {
				t.Error(title)
			}
			if !strings.Contains(title, tc.expectOutput) {
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
			rh, err := NewReceiver(&fakeClient{}, githubRepo, autoClose, resolvedLabel, extraLabels, tc.testName, tc.labelsTmpl)
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
			if tc.expectOutput == "" && labels != "" {
				t.Error(rh.TitleTmpl)
			}
			if !strings.Contains(labels, tc.expectOutput) {
				t.Error(rh.TitleTmpl)
			}
		})
	}
}


