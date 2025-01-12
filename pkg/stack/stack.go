package stack

import (
	"fmt"
)

// Stack defines a stack of integers
type Stack struct {
	items []int
}

// Push adds an element to the stack
func (s *Stack) Push(item int) {
	s.items = append(s.items, item)
}

// Pop removes and returns the last element from the stack (LIFO)
func (s *Stack) PopLast() (int, error) {
	if len(s.items) == 0 {
		return 0, fmt.Errorf("stack is empty")
	}
	lastIndex := len(s.items) - 1
	item := s.items[lastIndex]
	s.items = s.items[:lastIndex]
	return item, nil
}

// Pop removes and returns the first element from the stack (FIFO)
func (s *Stack) PopFirst() (int, error) {
	if len(s.items) == 0 {
		return 0, fmt.Errorf("stack is empty")
	}
	item := s.items[0]
	s.items = s.items[1:]
	return item, nil
}

// Peek returns the last element without removing it
func (s *Stack) Peek() (int, error) {
	if len(s.items) == 0 {
		return 0, fmt.Errorf("stack is empty")
	}
	return s.items[len(s.items)-1], nil
}

// Get returns the element at the specified index
func (s *Stack) Get(index int) (int, error) {
	if index < 0 || index >= len(s.items) {
		return 0, fmt.Errorf("index out of range")
	}
	return s.items[index], nil
}

// IsEmpty checks if the stack is empty
func (s *Stack) IsEmpty() bool {
	return len(s.items) == 0
}

// Size returns the number of elements in the stack
func (s *Stack) Size() int {
	return len(s.items)
}
