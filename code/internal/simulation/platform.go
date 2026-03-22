package simulation

import (
	"sync"
	"time"

	"swarm-predict/internal/model"

	"github.com/google/uuid"
)

// Platform is the simulated social media environment.
type Platform struct {
	mu    sync.RWMutex
	posts []model.Post
}

func NewPlatform() *Platform {
	return &Platform{}
}

func (p *Platform) CreatePost(authorID, content string) *model.Post {
	p.mu.Lock()
	defer p.mu.Unlock()
	post := model.Post{
		ID:        uuid.New().String(),
		AuthorID:  authorID,
		Content:   content,
		Timestamp: time.Now(),
	}
	p.posts = append(p.posts, post)
	return &post
}

func (p *Platform) Like(postID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := range p.posts {
		if p.posts[i].ID == postID {
			p.posts[i].Likes++
			return
		}
	}
}

func (p *Platform) Repost(postID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := range p.posts {
		if p.posts[i].ID == postID {
			p.posts[i].Reposts++
			return
		}
	}
}

func (p *Platform) Reply(postID, authorID, content string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := range p.posts {
		if p.posts[i].ID == postID {
			p.posts[i].Replies = append(p.posts[i].Replies, model.Reply{
				ID:        uuid.New().String(),
				AuthorID:  authorID,
				Content:   content,
				Timestamp: time.Now(),
			})
			return
		}
	}
}

// GetFeed returns recent posts, optionally limited.
func (p *Platform) GetFeed(limit int) []model.Post {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.posts) <= limit {
		result := make([]model.Post, len(p.posts))
		copy(result, p.posts)
		return result
	}
	result := make([]model.Post, limit)
	copy(result, p.posts[len(p.posts)-limit:])
	return result
}

func (p *Platform) GetAllPosts() []model.Post {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]model.Post, len(p.posts))
	copy(result, p.posts)
	return result
}

func (p *Platform) PostCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.posts)
}

// Execute applies a decision to the platform.
func (p *Platform) Execute(d model.Decision) {
	switch d.ActionType {
	case "post":
		p.CreatePost(d.AgentID, d.Content)
	case "like":
		p.Like(d.TargetID)
	case "repost":
		p.Repost(d.TargetID)
	case "reply":
		p.Reply(d.TargetID, d.AgentID, d.Content)
	}
}
