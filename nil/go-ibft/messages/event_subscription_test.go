package messages

import (
	"testing"

	"github.com/NilFoundation/nil/nil/go-ibft/messages/proto"
	"github.com/stretchr/testify/assert"
)

func TestEventSubscription_EventSupported(t *testing.T) {
	t.Parallel()

	type signalDetails struct {
		messageType   proto.MessageType
		view          *proto.View
		totalMessages int
	}

	commonDetails := SubscriptionDetails{
		MessageType: proto.MessageType_PREPARE,
		View: &proto.View{
			Height: 0,
			Round:  0,
		},
		MinNumMessages: 10,
	}

	testTable := []struct {
		name                string
		subscriptionDetails SubscriptionDetails
		event               signalDetails
		shouldSupport       bool
	}{
		{
			"Same signal as subscription",
			commonDetails,
			signalDetails{
				commonDetails.MessageType,
				commonDetails.View,
				commonDetails.MinNumMessages,
			},
			true,
		},
		{
			"Message round > round than subscription (supported)",
			SubscriptionDetails{
				MessageType:    commonDetails.MessageType,
				View:           commonDetails.View,
				MinNumMessages: commonDetails.MinNumMessages,
				HasMinRound:    true,
			},
			signalDetails{
				commonDetails.MessageType,
				&proto.View{
					Height: commonDetails.View.GetHeight(),
					Round:  commonDetails.View.GetRound() + 1,
				},
				commonDetails.MinNumMessages,
			},
			true,
		},
		{
			"Message round == round than subscription (supported)",
			SubscriptionDetails{
				MessageType:    commonDetails.MessageType,
				View:           commonDetails.View,
				MinNumMessages: commonDetails.MinNumMessages,
				HasMinRound:    true,
			},
			signalDetails{
				commonDetails.MessageType,
				commonDetails.View,
				commonDetails.MinNumMessages,
			},
			true,
		},
		{
			"Message round > round than subscription (not supported)",
			commonDetails,
			signalDetails{
				commonDetails.MessageType,
				&proto.View{
					Height: commonDetails.View.GetHeight(),
					Round:  commonDetails.View.GetRound() + 1,
				},
				commonDetails.MinNumMessages,
			},
			false,
		},
		{
			"Message round < round than subscription (not supported)",
			SubscriptionDetails{
				MessageType: commonDetails.MessageType,
				View: &proto.View{
					Height: commonDetails.View.GetHeight(),
					Round:  commonDetails.View.GetRound() + 10,
				},
				MinNumMessages: commonDetails.MinNumMessages,
				HasMinRound:    true,
			},
			signalDetails{
				commonDetails.MessageType,
				&proto.View{
					Height: commonDetails.View.GetHeight(),
					Round:  commonDetails.View.GetRound() + 10 - 1,
				},
				commonDetails.MinNumMessages,
			},
			false,
		},
		{
			"Invalid message type",
			commonDetails,
			signalDetails{
				proto.MessageType_COMMIT,
				commonDetails.View,
				commonDetails.MinNumMessages,
			},
			false,
		},
		{
			"Invalid message height",
			commonDetails,
			signalDetails{
				commonDetails.MessageType,
				&proto.View{
					Height: commonDetails.View.GetHeight() + 1,
					Round:  commonDetails.View.GetRound(),
				},
				commonDetails.MinNumMessages,
			},
			false,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			subscription := &eventSubscription{
				details:  testCase.subscriptionDetails,
				outputCh: make(chan uint64, 1),
				notifyCh: make(chan uint64, 1),
				doneCh:   make(chan struct{}),
			}

			t.Cleanup(func() {
				subscription.close()
			})

			event := testCase.event

			assert.Equal(
				t,
				testCase.shouldSupport,
				subscription.eventSupported(
					event.messageType,
					event.view,
				),
			)
		})
	}
}
