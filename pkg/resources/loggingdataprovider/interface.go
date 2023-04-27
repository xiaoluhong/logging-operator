// Copyright © 2021 Cisco Systems, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package loggingdataprovider

import (
	"context"

	"github.com/kube-logging/logging-operator/pkg/sdk/logging/api/v1beta1"
)

type LoggingDataProvider interface {
	GetReplicaCount(ctx context.Context, logging *v1beta1.Logging) (*int32, error)
}