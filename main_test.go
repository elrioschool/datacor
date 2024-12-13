// FILE: main_test.go

package donationsbystudent

import (
	"testing"
)

func TestAddParent(t *testing.T) {
	tests := []struct {
		name       string
		student    Student
		parentName string
		expected   Student
	}{
		{
			name:       "Add first parent",
			student:    Student{},
			parentName: "Parent1",
			expected:   Student{Parent1: "Parent1"},
		},
		{
			name:       "Add second parent",
			student:    Student{Parent1: "Parent1"},
			parentName: "Parent2",
			expected:   Student{Parent1: "Parent1", Parent2: "Parent2"},
		},
		{
			name:       "Add third parent",
			student:    Student{Parent1: "Parent1", Parent2: "Parent2"},
			parentName: "Parent3",
			expected:   Student{Parent1: "Parent1", Parent2: "Parent2", Parent3: "Parent3"},
		},
		{
			name:       "No space for new parent",
			student:    Student{Parent1: "Parent1", Parent2: "Parent2", Parent3: "Parent3"},
			parentName: "Parent4",
			expected:   Student{Parent1: "Parent1", Parent2: "Parent2", Parent3: "Parent3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addParent(&tt.student, tt.parentName)
			if tt.student != tt.expected {
				t.Errorf("got %v, want %v", tt.student, tt.expected)
			}
		})
	}
}
