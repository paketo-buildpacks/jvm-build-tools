// Copyright 2018-2026 the original author or authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/paketo-buildpacks/packit/v2/cargo"
)

var httpClient = &http.Client{}

func defaultTargets() []cargo.ConfigTarget {
	return []cargo.ConfigTarget{
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "arm64"},
	}
}

func fetchBody(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch %s\n%w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to fetch %s: status code %d", url, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func fetchJSON(url string, v any) error {
	body, err := fetchBody(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("unable to parse %s\n%w", url, err)
	}

	return nil
}
