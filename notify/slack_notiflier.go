/**
 * Copyright (c) 2019-present Future Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package notify

import (
	"context"
	"io"
	"fmt"
	"os"
	"strconv"
	"strings"
	"log"

	"github.com/slack-go/slack"
	chart "github.com/wcharczuk/go-chart/v2"
)

type slackNotifier struct {
	slackAPIToken string
	slackChannel  string
}

func NewSlackNotifier(slackAPIToken, slackChannel string) *slackNotifier {
	return &slackNotifier{
		slackAPIToken: slackAPIToken,
		slackChannel:  slackChannel,
	}
}

func (n *slackNotifier) PostImage(ctx context.Context, r io.Reader) error {
	_, err := slack.New(n.slackAPIToken).UploadFileContext(ctx,
		slack.FileUploadParameters{
			Reader:   r,
			Filename: "Stacked Bar Chart on Projects",
			Channels: []string{n.slackChannel},
		})
	return err
}

func (n* slackNotifier) PostMessage(ctx context.Context, msg string) error {
	_, _, err := slack.New(n.slackAPIToken).PostMessageContext(ctx,
		n.slackChannel,
		slack.MsgOptionText(msg, false),
	)
	return err
}

func (n* slackNotifier) NotifyByProject(ctx context.Context, values []chart.Value) error {
	for _, v := range values {
		env := "MAX_" + strings.Replace(v.Label, "-", "_", -1)
		threshold, exist := os.LookupEnv(env)
		if !exist {
			log.Printf("%s is not defined", env)
			continue
		}
		thresholdInt, _ := strconv.Atoi(threshold)
		value := int(v.Value)
		if value > thresholdInt {
			msg := fmt.Sprintf("<!channel> %s %d is over %d", v.Label, value, thresholdInt)
			if v.Label == "total" {
				msg = fmt.Sprintf("<!channel> you must to review current invoice which is higher than %d€: %d", value, thresholdInt)
			}
			err := n.PostMessage(ctx, msg)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
