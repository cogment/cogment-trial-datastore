// Copyright 2021 AI Redefined Inc. <dev+cogment@ai-r.com>
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

package backend

import (
	"testing"
	"time"

	grpcapi "github.com/cogment/cogment-trial-datastore/grpcapi/cogment/api"
	"github.com/openlyinc/pointy"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

var trialParams *grpcapi.TrialParams = &grpcapi.TrialParams{
	TrialConfig: &grpcapi.TrialConfig{
		Content: []byte("a trial config"),
	},
	MaxSteps:      12,
	MaxInactivity: 600,
	Environment: &grpcapi.EnvironmentParams{
		Endpoint:       "grpc://environment:9000",
		Implementation: "my-environment-implementation",
		Config: &grpcapi.EnvironmentConfig{
			Content: []byte("an environment config"),
		},
	},
	Actors: []*grpcapi.ActorParams{
		{
			Name:           "my-actor-1",
			ActorClass:     "my-actor-class-1",
			Endpoint:       "grpc://actor:9000",
			Implementation: "my-actor-implementation",
			Config: &grpcapi.ActorConfig{
				Content: []byte("an actor config"),
			},
		},
		{
			Name:           "my-actor-2",
			ActorClass:     "my-actor-class-2",
			Endpoint:       "grpc://actor:9000",
			Implementation: "my-actor-implementation",
			Config: &grpcapi.ActorConfig{
				Content: []byte("another actor config"),
			},
		},
	},
}

var trialSample1 *grpcapi.TrialSample = &grpcapi.TrialSample{
	UserId:    "my-user-id",
	TrialId:   "my-trial",
	TickId:    12,
	Timestamp: uint64(time.Now().Unix()),
	ActorSamples: []*grpcapi.TrialActorSample{
		{
			Actor:       0,
			Observation: pointy.Uint32(0),
			Action:      pointy.Uint32(1),
			Reward:      pointy.Float32(0.5),
			ReceivedRewards: []*grpcapi.TrialActorSampleReward{
				{
					Sender:     -1,
					Reward:     0.5,
					Confidence: 1,
					UserData:   nil,
				},
				{
					Sender:     1,
					Reward:     0.5,
					Confidence: 0.2,
					UserData:   pointy.Uint32(2),
				},
			},
		},
		{
			Actor:       1,
			Observation: pointy.Uint32(0),
			Action:      pointy.Uint32(3),
			SentRewards: []*grpcapi.TrialActorSampleReward{
				{
					Receiver:   0,
					Reward:     0.5,
					Confidence: 0.2,
					UserData:   pointy.Uint32(2),
				},
			},
			ReceivedMessages: []*grpcapi.TrialActorSampleMessage{
				{
					Sender:  -1,
					Payload: 4,
				},
			},
			SentMessages: []*grpcapi.TrialActorSampleMessage{
				{
					Receiver: -1,
					Payload:  5,
				},
			},
		},
	},
	Payloads: [][]byte{
		[]byte("an observation"),
		[]byte("an action"),
		[]byte("a reward user data"),
		[]byte("another action"),
		[]byte("a message payload"),
		[]byte("another message payload"),
	},
}

func TestNoFilters(t *testing.T) {
	filteredTrialSample1 := filterTrialSample(trialSample1, make(trialActorFilter), make(sampleFieldsFilter))
	assert.True(t, proto.Equal(trialSample1, filteredTrialSample1))
}

func TestFieldFiltersFilterOutRewardsAndMessages(t *testing.T) {
	fieldFilter := createSampleFieldsFilter([]grpcapi.TrialSampleField{
		grpcapi.TrialSampleField_TRIAL_SAMPLE_FIELD_ACTION,
		grpcapi.TrialSampleField_TRIAL_SAMPLE_FIELD_OBSERVATION,
		grpcapi.TrialSampleField_TRIAL_SAMPLE_FIELD_REWARD,
	})
	filteredTrialSample1 := filterTrialSample(trialSample1, make(trialActorFilter), fieldFilter)

	assert.False(t, proto.Equal(trialSample1, filteredTrialSample1))
	assert.Less(t, proto.Size(filteredTrialSample1), proto.Size(trialSample1))
	t.Logf("Full size=%vB", proto.Size(trialSample1))
	t.Logf("Filtered size=%vB", proto.Size(filteredTrialSample1))

	assert.Len(t, filteredTrialSample1.ActorSamples[0].ReceivedRewards, 0)
	assert.Len(t, filteredTrialSample1.ActorSamples[0].SentRewards, 0)
	assert.Len(t, filteredTrialSample1.ActorSamples[0].ReceivedMessages, 0)
	assert.Len(t, filteredTrialSample1.ActorSamples[0].SentMessages, 0)
	assert.Len(t, filteredTrialSample1.ActorSamples[1].ReceivedRewards, 0)
	assert.Len(t, filteredTrialSample1.ActorSamples[1].SentRewards, 0)
	assert.Len(t, filteredTrialSample1.ActorSamples[1].ReceivedMessages, 0)
	assert.Len(t, filteredTrialSample1.ActorSamples[1].SentMessages, 0)

	assert.NotEmpty(t, filteredTrialSample1.Payloads[0])
	assert.NotEmpty(t, filteredTrialSample1.Payloads[1])
	assert.Empty(t, filteredTrialSample1.Payloads[2])
	assert.NotEmpty(t, filteredTrialSample1.Payloads[3])
	assert.Empty(t, filteredTrialSample1.Payloads[4])
	assert.Empty(t, filteredTrialSample1.Payloads[5])

	twiceFilteredTrialSample1 := filterTrialSample(filteredTrialSample1, make(trialActorFilter), fieldFilter)

	assert.True(t, proto.Equal(twiceFilteredTrialSample1, filteredTrialSample1))
}

func TestFieldFiltersFilterOutReward(t *testing.T) {
	fieldFilter := createSampleFieldsFilter([]grpcapi.TrialSampleField{
		grpcapi.TrialSampleField_TRIAL_SAMPLE_FIELD_ACTION,
		grpcapi.TrialSampleField_TRIAL_SAMPLE_FIELD_OBSERVATION,
		grpcapi.TrialSampleField_TRIAL_SAMPLE_FIELD_RECEIVED_REWARDS,
		grpcapi.TrialSampleField_TRIAL_SAMPLE_FIELD_SENT_REWARDS,
		grpcapi.TrialSampleField_TRIAL_SAMPLE_FIELD_RECEIVED_MESSAGES,
		grpcapi.TrialSampleField_TRIAL_SAMPLE_FIELD_SENT_MESSAGES,
	})
	filteredTrialSample1 := filterTrialSample(trialSample1, make(trialActorFilter), fieldFilter)

	assert.False(t, proto.Equal(trialSample1, filteredTrialSample1))
	assert.Less(t, proto.Size(filteredTrialSample1), proto.Size(trialSample1))
	t.Logf("Full size=%vB", proto.Size(trialSample1))
	t.Logf("Filtered size=%vB", proto.Size(filteredTrialSample1))

	assert.Nil(t, filteredTrialSample1.ActorSamples[0].Reward)
	assert.Nil(t, filteredTrialSample1.ActorSamples[1].Reward)
}

func TestActorNameFilters(t *testing.T) {
	actorNameFilter := createActorFilter(createFilterFromStringArray([]string{"my-actor-2"}), createFilterFromStringArray([]string{}), createFilterFromStringArray([]string{}), trialParams)
	filteredTrialSample1 := filterTrialSample(trialSample1, actorNameFilter, make(sampleFieldsFilter))

	assert.False(t, proto.Equal(trialSample1, filteredTrialSample1))
	assert.Less(t, proto.Size(filteredTrialSample1), proto.Size(trialSample1))

	assert.Len(t, filteredTrialSample1.ActorSamples, 1)

	assert.NotEmpty(t, filteredTrialSample1.Payloads[0])
	assert.Empty(t, filteredTrialSample1.Payloads[1])
	assert.NotEmpty(t, filteredTrialSample1.Payloads[2])
	assert.NotEmpty(t, filteredTrialSample1.Payloads[3])
	assert.NotEmpty(t, filteredTrialSample1.Payloads[4])
	assert.NotEmpty(t, filteredTrialSample1.Payloads[5])

	twiceFilteredTrialSample1 := filterTrialSample(filteredTrialSample1, actorNameFilter, make(sampleFieldsFilter))

	assert.True(t, proto.Equal(twiceFilteredTrialSample1, filteredTrialSample1))
}