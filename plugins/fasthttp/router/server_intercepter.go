// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package router

import (
	"fmt"

	"github.com/apache/skywalking-go/plugins/core/operator"
	"github.com/apache/skywalking-go/plugins/core/tracing"
	"github.com/valyala/fasthttp"
)

type ServerInterceptor struct {
}

func (h *ServerInterceptor) BeforeInvoke(invocation operator.Invocation) error {
	ctx := invocation.Args()[0].(*fasthttp.RequestCtx)
	s, err := tracing.CreateEntrySpan(fmt.Sprintf("%s:%s", string(ctx.Method()), ctx.URI().String()),
		func(headerKey string) (string, error) {
			return string(ctx.Request.Header.Peek(headerKey)), nil
		}, tracing.WithLayer(tracing.SpanLayerHTTP),
		tracing.WithTag(tracing.TagHTTPMethod, string(ctx.Method())),
		tracing.WithTag(tracing.TagURL, ctx.URI().String()),
		tracing.WithComponent(5020))
	if err != nil {
		return err
	}

	invocation.SetContext(s)
	return nil
}

func (h *ServerInterceptor) AfterInvoke(invocation operator.Invocation, result ...interface{}) error {
	if invocation.GetContext() == nil {
		return nil
	}
	span := invocation.GetContext().(tracing.Span)
	if ctx, ok := invocation.Args()[0].(*fasthttp.RequestCtx); ok {
		if ctx.Response.StatusCode() >= 400 {
			span.Error()
		}
		span.Tag(tracing.TagStatusCode, fmt.Sprintf("%d", ctx.Response.StatusCode()))
	}
	span.End()
	return nil
}
