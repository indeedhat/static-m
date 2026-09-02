package main

import (
	"bytes"
	"errors"
	"html/template"
	"mime"
	"net/url"
	"os"
	"path"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	ScorePathWildcard    = 1e6
	ScoreAdditionalQuery = 1e4
	ScoreQueryWildcard   = 1
)

var templateFuncs = template.FuncMap{
	"last_idx": func(list []string) int { return len(list) - 1 },
}

type ResponseConfig struct {
	Status  int               `yaml:"status"`
	Headers map[string]string `yaml:"headers"`
}

type Document struct {
	// Header values
	Mime       string          `yaml:"mime"`
	Path       string          `yaml:"path"`
	Method     string          `yaml:"method"`
	IsTemplate bool            `yaml:"template"`
	Response   *ResponseConfig `yaml:"response"`

	// internal state
	url  *url.URL
	body []byte
	info os.FileInfo
}

func NewDocument(filePath string, data []byte, info os.FileInfo) (*Document, error) {
	sep := []byte("\n---\n")
	idx := bytes.Index(data, sep)
	if idx == -1 {
		return nil, errors.New("missing header")
	}

	headerData := data[:idx]
	bodyData := data[idx+len(sep):]

	var d Document
	var err error
	if err = yaml.Unmarshal(headerData, &d); err != nil {
		return nil, err
	}

	if d.Mime == "" {
		d.Mime = mime.TypeByExtension(path.Ext(filePath))
	}

	if d.Path == "" {
		return nil, errors.New("invalid header, must have path")
	}

	d.info = info
	d.body = bodyData

	if d.url, err = url.Parse(d.Path); err != nil {
		return nil, err
	}

	return &d, nil
}

func (d Document) Render(args map[string]any) ([]byte, error) {
	if !d.IsTemplate {
		return d.body, nil
	}

	t, err := template.New(d.Path).Funcs(templateFuncs).Parse(string(d.body))
	if err != nil {
		return nil, err
	}

	var b []byte
	buf := bytes.NewBuffer(b)
	if err = t.Execute(buf, args); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (d Document) Match(u *url.URL, method string) (match bool, score uint, data map[string]any) {
	if d.Method != "" && d.Method != method {
		return false, 0, nil
	}

	if u.String() == d.url.String() {
		return true, 0, nil
	}

	docPath := strings.Split(d.url.Path, "/")
	subjectPath := strings.Split(u.Path, "/")

	if len(docPath) > len(subjectPath) {
		return false, 0, nil
	}

	data = make(map[string]any)

	pMatch, pScore := d.matchPath(docPath, subjectPath, data)
	if !pMatch {
		return false, 0, nil
	}

	qMatch, qScore := d.matchQuery(d.url.Query(), u.Query(), data)
	if !qMatch {
		return false, 0, nil
	}

	return true, pScore + qScore, data
}

func (d Document) matchPath(doc, subject []string, data map[string]any) (math bool, score uint) {
	if len(doc) == 0 {
		if len(subject) == 0 {
			return true, 0
		}

		return false, 0
	}

	var curI int
	cur := doc[curI]

	for i, part := range subject {
		if cur == part {
			curI++
			if curI >= len(doc) {
				if i == len(subject)-1 {
					break
				}

				return false, 0
			}

			cur = doc[curI]
			continue
		}

		wild, greedy, name := parseWildcard(cur)
		if !wild {
			return false, 0
		}

		score += ScorePathWildcard

		if !greedy {
			data[name] = part
			curI++
			if curI >= len(doc) {
				if i == len(subject)-1 {
					break
				}

				return false, 0
			}

			cur = doc[curI]
			continue
		}

		if v, found := data[name]; found {
			data[name] = v.(string) + "/" + part
		} else {
			data[name] = "/" + part
		}

		if i == len(subject)-1 {
			curI++
		}
	}

	if curI < len(doc) {
		return false, 0
	}

	return true, score
}

func (d Document) matchQuery(doc, subject url.Values, data map[string]any) (math bool, score uint) {
	for k, v := range doc {
		if !subject.Has(k) {
			return false, 0
		}

		subs := subject[k]

		wild, greedy, name := parseWildcard(v[0])
		if name == "" {
			name = k
		}

		if wild && !greedy {
			if len(subs) > 1 {
				return false, 0
			}

			score += ScoreQueryWildcard
			data[name] = subs[0]
			continue
		}

		if greedy {
			score += uint(ScoreQueryWildcard * len(subs))
			data[name] = subs
			continue
		}

		if !slicesHaveSameValues(subs, v) {
			return false, 0
		}
	}

	score += uint((len(subject) - len(doc)) * ScoreAdditionalQuery)

	return true, score
}

func slicesHaveSameValues(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for _, v := range a {
		if !slices.Contains(b, v) {
			return false
		}
	}

	return true
}
