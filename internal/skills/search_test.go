package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple text",
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "with numbers",
			input:    "test123 abc",
			expected: []string{"test123", "abc"},
		},
		{
			name:     "punctuation removed",
			input:    "hello, world!",
			expected: []string{"hello", "world"},
		},
		{
			name:     "single char filtered",
			input:    "a b c",
			expected: nil,
		},
		{
			name:     "mixed case",
			input:    "Hello World",
			expected: []string{"hello", "world"},
		},
		{
			name:     "special chars",
			input:    "foo@bar.com test",
			expected: []string{"foo", "bar", "com", "test"},
		},
		{
			name:     "unicode",
			input:    "你好世界",
			expected: []string{"你好世界"}, // unicode passes through
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewIndex(t *testing.T) {
	idx := NewIndex()

	assert.NotNil(t, idx)
	assert.NotNil(t, idx.df)
	assert.Equal(t, 1.2, idx.k1)
	assert.Equal(t, 0.75, idx.b)
}

func TestIndexBuild(t *testing.T) {
	idx := NewIndex()

	skills := []Info{
		{Name: "Docker", Description: "Container management and orchestration"},
		{Name: "Kubernetes", Description: "Container orchestration platform"},
		{Name: "Python", Description: "Programming language for data science"},
	}

	idx.Build(skills)

	assert.Len(t, idx.docs, 3)
	assert.True(t, idx.avgDL > 0)
}

func TestIndexSearch(t *testing.T) {
	idx := NewIndex()

	skills := []Info{
		{Name: "Docker", Description: "Container management and orchestration"},
		{Name: "Kubernetes", Description: "Container orchestration platform"},
		{Name: "Python", Description: "Programming language for data science"},
		{Name: "Git", Description: "Version control system"},
	}
	idx.Build(skills)

	tests := []struct {
		name         string
		query        string
		maxResults   int
		expectCount  int
		expectFirst  string
	}{
		{
			name:        "search container",
			query:       "container",
			maxResults:  5,
			expectCount: 2,
			expectFirst: "Kubernetes", // scores higher due to more term matches
		},
		{
			name:        "search python",
			query:       "python",
			maxResults:  5,
			expectCount: 1,
			expectFirst: "Python",
		},
		{
			name:        "limit results",
			query:       "container",
			maxResults:  1,
			expectCount: 1,
			expectFirst: "Kubernetes",
		},
		{
			name:        "no results",
			query:       "java",
			maxResults:  5,
			expectCount: 0,
		},
		{
			name:        "empty query",
			query:       "",
			maxResults:  5,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := idx.Search(tt.query, tt.maxResults)
			assert.Len(t, results, tt.expectCount)
			if tt.expectCount > 0 {
				assert.Equal(t, tt.expectFirst, results[0].Name)
			}
		})
	}
}

func TestIndexSearchSorting(t *testing.T) {
	idx := NewIndex()

	skills := []Info{
		{Name: "Python", Description: "Programming language Python"},
		{Name: "PyTorch", Description: "Machine learning framework PyTorch"},
		{Name: "Pytest", Description: "Testing framework for Python"},
	}
	idx.Build(skills)

	results := idx.Search("python", 5)

	// All results should have scores > 0
	for _, r := range results {
		assert.True(t, r.Score > 0)
	}

	// Results should be sorted by score descending
	for i := 1; i < len(results); i++ {
		assert.GreaterOrEqual(t, results[i-1].Score, results[i].Score)
	}
}

func TestSkillSearchResultFields(t *testing.T) {
	idx := NewIndex()

	skills := []Info{
		{
			Name:        "Docker",
			Slug:        "docker",
			Description: "Container management",
			Path:        "/skills/docker/SKILL.md",
			BaseDir:     "/skills/docker",
			Source:      "workspace",
		},
	}
	idx.Build(skills)

	results := idx.Search("docker", 1)

	assert.Len(t, results, 1)
	r := results[0]
	assert.Equal(t, "Docker", r.Name)
	assert.Equal(t, "docker", r.Slug)
	assert.Equal(t, "Container management", r.Description)
	assert.Equal(t, "/skills/docker/SKILL.md", r.Location)
	assert.Equal(t, "/skills/docker", r.BaseDir)
	assert.Equal(t, "workspace", r.Source)
}

func TestIndexEmpty(t *testing.T) {
	idx := NewIndex()
	idx.Build([]Info{})

	results := idx.Search("test", 5)
	assert.Nil(t, results)
}

func TestSortScored(t *testing.T) {
	results := []scored{
		{score: 1.0},
		{score: 3.0},
		{score: 2.0},
	}

	sortScored(results)

	assert.Equal(t, 3.0, results[0].score)
	assert.Equal(t, 2.0, results[1].score)
	assert.Equal(t, 1.0, results[2].score)
}

func TestIndexSearchMaxResultsZero(t *testing.T) {
	idx := NewIndex()

	skills := []Info{
		{Name: "Docker", Description: "Container management"},
		{Name: "Kubernetes", Description: "Container orchestration"},
	}
	idx.Build(skills)

	// Should default to 5 when maxResults <= 0
	results := idx.Search("container", 0)
	assert.Len(t, results, 2)
}
