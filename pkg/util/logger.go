/*
 * Copyright 2025 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"github.com/SENERGY-Platform/go-service-base/structured-logger"
	"github.com/SENERGY-Platform/go-service-base/structured-logger/attributes"
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/config"
	"io"
	"log/slog"
)

var Logger *slog.Logger

func InitLogger(c config.LoggerConfig, out io.Writer, organization, project string) {
	recordTime := structured_logger.NewRecordTime(c.TimeFormat, c.TimeUtc)
	options := &slog.HandlerOptions{
		AddSource:   c.AddSource,
		Level:       structured_logger.GetLevel(c.Level, slog.LevelInfo),
		ReplaceAttr: recordTime.ReplaceAttr,
	}
	handler := structured_logger.GetHandler(c.Handler, out, options, slog.Default().Handler())
	var attr []slog.Attr
	if c.AddMeta {
		if organization != "" {
			attr = append(attr, slog.String(attributes.OrganizationKey, organization))
		}
		if project != "" {
			attr = append(attr, slog.String(attributes.ProjectKey, project))
		}
	}
	handler = handler.WithAttrs(attr)
	Logger = slog.New(handler)
}
