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
	"context"
	"github.com/SENERGY-Platform/api-docs-provider/pkg/models"
)

func GetReqID(ctx context.Context) string {
	val := ctx.Value(models.ContextRequestID)
	if val != nil {
		if str, ok := val.(string); ok {
			return str + " "
		}
	}
	return ""
}
