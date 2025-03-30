// Copyright 2025 Kristopher Rahim Afful-Brown. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nettest

import (
	"bytes"
	"context"
	"encoding/json"
	"os"

	"github.com/Shopify/toxiproxy/v2"
	"github.com/Shopify/toxiproxy/v2/toxics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

// Proxy
type Proxy struct {
	p *toxiproxy.Proxy
	s *toxiproxy.ApiServer
}

// Listen
func (p *Proxy) Listen() string {
	if p.p == nil {
		return ""
	}
	return p.p.Listen
}

// AddToxic
func (p *Proxy) AddToxic(typ string, upstream bool, toxic toxics.Toxic) (string, error) {
	var stream string
	if upstream {
		stream = "upstream"
	} else {
		stream = "downstream"
	}
	b, err := json.Marshal(map[string]any{
		"name":       typ + "_" + stream,
		"type":       typ,
		"stream":     stream,
		"attributes": toxic,
	})
	if err != nil {
		return "", err
	}

	res, err := p.p.Toxics.AddToxicJson(bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	return res.Name, nil
}

// RemoveToxic
func (p *Proxy) RemoveToxic(name string) error {
	err := p.p.Toxics.RemoveToxic(context.Background(), name)
	if err != nil {
		return err
	}
	return nil
}

// Close
func (p *Proxy) Close() error {
	if p.p != nil {
		p.p.Stop()
	}
	if p.s != nil {
		return p.s.Shutdown()
	}
	return nil
}

// NewProxy returns a new [Proxy]
func NewProxy(name, upstream string) *Proxy {
	xm := toxiproxy.NewMetricsContainer(prometheus.NewRegistry())
	xl := zerolog.New(os.Stderr).Level(zerolog.ErrorLevel)
	s := toxiproxy.NewServer(xm, xl)
	proxy := toxiproxy.NewProxy(s, name, "localhost:0", upstream)
	proxy.Start()
	return &Proxy{p: proxy, s: s}
}
