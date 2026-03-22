package simulation

import (
	"math"
	"sort"
	"strings"
	"time"

	"swarm-predict/internal/model"
)

// FeedAlgorithm generates a personalized feed for each agent.
type FeedAlgorithm struct {
	platform *Platform
}

func NewFeedAlgorithm(platform *Platform) *FeedAlgorithm {
	return &FeedAlgorithm{platform: platform}
}

// PersonalizedFeed returns a ranked feed for the given agent, considering
// recency, engagement (likes/reposts), and relevance to the agent's role.
func (f *FeedAlgorithm) PersonalizedFeed(agent model.AgentProfile, relations []model.SocialRelation, limit int) []model.Post {
	posts := f.platform.GetAllPosts()
	if len(posts) == 0 {
		return nil
	}

	followedSet := make(map[string]float64)
	for _, rel := range relations {
		if rel.Likability > 0 {
			followedSet[rel.TargetID] = rel.Likability
		}
	}

	type scored struct {
		post  model.Post
		score float64
	}

	now := time.Now()
	scoredPosts := make([]scored, 0, len(posts))

	for _, post := range posts {
		if post.AuthorID == agent.ID {
			continue // skip own posts
		}

		recency := 1.0 / (1.0 + now.Sub(post.Timestamp).Hours())
		engagement := math.Log2(float64(2 + post.Likes + post.Reposts*2))
		socialBoost := 1.0
		if boost, ok := followedSet[post.AuthorID]; ok {
			socialBoost = 1.0 + boost
		}
		relevance := topicRelevance(post.Content, agent.Role, agent.Personality)

		score := recency*0.3 + engagement*0.2 + socialBoost*0.3 + relevance*0.2
		scoredPosts = append(scoredPosts, scored{post: post, score: score})
	}

	sort.Slice(scoredPosts, func(i, j int) bool {
		return scoredPosts[i].score > scoredPosts[j].score
	})

	if limit > len(scoredPosts) {
		limit = len(scoredPosts)
	}
	result := make([]model.Post, limit)
	for i := 0; i < limit; i++ {
		result[i] = scoredPosts[i].post
	}
	return result
}

// topicRelevance is a simple keyword-overlap heuristic.
func topicRelevance(content, role, personality string) float64 {
	contentWords := strings.Fields(strings.ToLower(content))
	keywords := strings.Fields(strings.ToLower(role + " " + personality))

	if len(keywords) == 0 || len(contentWords) == 0 {
		return 0
	}

	keywordSet := make(map[string]bool)
	for _, kw := range keywords {
		if len(kw) > 3 {
			keywordSet[kw] = true
		}
	}

	matches := 0
	for _, w := range contentWords {
		if keywordSet[w] {
			matches++
		}
	}
	return float64(matches) / float64(len(keywordSet)+1)
}
