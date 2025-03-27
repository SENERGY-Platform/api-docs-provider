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
	"github.com/SENERGY-Platform/swagger-docs-provider/pkg/util/slog_attr"
	"os"
	"os/signal"
)

func WaitForSignal(ctx context.Context, signals ...os.Signal) {
	ch := make(chan os.Signal, 1)
	for _, sig := range signals {
		signal.Notify(ch, sig)
	}
	select {
	case sig := <-ch:
		Logger.Warn("caught os signal", slog_attr.NumberKey, sig.String())
		break
	case <-ctx.Done():
		break
	}
	signal.Stop(ch)
}
