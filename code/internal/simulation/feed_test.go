package simulation

import (
	"testing"

	"swarm-predict/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestFeedAlgorithm_PersonalizedFeed(t *testing.T) {
	p := NewPlatform()
	p.CreatePost("gov-1", "New economic reform policy announced today")
	p.CreatePost("media-1", "Breaking: analysts predict market impact")
	p.CreatePost("pub-1", "I love this sunny weather today")
	p.CreatePost("analyst-1", "Data shows economy growing at 3% rate")

	feed := NewFeedAlgorithm(p)

	agent := model.AgentProfile{
		ID:          "pub-2",
		Role:        "analyst",
		Personality: "data-driven economy",
	}

	result := feed.PersonalizedFeed(agent, nil, 3)
	assert.Len(t, result, 3)
	// Should not include own posts (agent is pub-2, not in posts)
	for _, post := range result {
		assert.NotEqual(t, "pub-2", post.AuthorID)
	}
}

func TestFeedAlgorithm_SkipsOwnPosts(t *testing.T) {
	p := NewPlatform()
	p.CreatePost("me", "my own post")
	p.CreatePost("other", "other post")

	feed := NewFeedAlgorithm(p)
	result := feed.PersonalizedFeed(model.AgentProfile{ID: "me"}, nil, 10)
	assert.Len(t, result, 1)
	assert.Equal(t, "other", result[0].AuthorID)
}

func TestFeedAlgorithm_SocialBoost(t *testing.T) {
	p := NewPlatform()
	p.CreatePost("friend", "friendly post")
	p.CreatePost("stranger", "stranger post")

	relations := []model.SocialRelation{
		{ObserverID: "me", TargetID: "friend", Likability: 0.9},
	}

	feed := NewFeedAlgorithm(p)
	result := feed.PersonalizedFeed(model.AgentProfile{ID: "me"}, relations, 2)
	assert.Len(t, result, 2)
	assert.Equal(t, "friend", result[0].AuthorID) // friend boosted to top
}

func TestFeedAlgorithm_Empty(t *testing.T) {
	p := NewPlatform()
	feed := NewFeedAlgorithm(p)
	result := feed.PersonalizedFeed(model.AgentProfile{ID: "x"}, nil, 5)
	assert.Nil(t, result)
}

func TestTopicRelevance(t *testing.T) {
	assert.Greater(t, topicRelevance("economy analyst data analysis report", "analyst", "data-driven economy"), 0.0)
	assert.Equal(t, 0.0, topicRelevance("", "analyst", ""))
	assert.Equal(t, 0.0, topicRelevance("hello world", "", ""))
}
