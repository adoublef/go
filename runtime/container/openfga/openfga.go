// Copyright Kristopher Rahim Afful-Brown 2025. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package openfga

import (
	"cmp"
	"context"
	"fmt"

	"github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/openfga"
)

const DefaultSecret = "openfga-secret"

type Container struct{ *openfga.OpenFGAContainer }

const DefaultImage = "openfga/openfga:v1.8.0" //"openfga/openfga:v1.8.16"

func Run(ctx context.Context, image string) (*Container, error) {
	c, err := openfga.Run(ctx,
		cmp.Or(image, DefaultImage),
		testcontainers.WithEnv(map[string]string{
			"OPENFGA_LOG_LEVEL":            "warn",
			"OPENFGA_AUTHN_METHOD":         "preshared",
			"OPENFGA_AUTHN_PRESHARED_KEYS": DefaultSecret,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create openfga instance: %w", err)
	}
	return &Container{c}, nil
}

const DefaultStoreId = "11111111111111111111111111"

func (c *Container) ConnectionClient(ctx context.Context, store string) (*client.OpenFgaClient, error) {
	uri, err := c.HttpEndpoint(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get http endpoint: %w", err)
	}
	fga, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl: uri,
		Credentials: &credentials.Credentials{
			Method: credentials.CredentialsMethodApiToken,
			Config: &credentials.Config{
				ApiToken: DefaultSecret,
			},
		},
		StoreId: store,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client endpoint: %w", err)
	}
	return fga, nil
}
