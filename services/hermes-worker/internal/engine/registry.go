package engine

import (
	"fmt"
	"sort"
)

type Registry struct {
	executors map[string]ActionExecutor
}

func NewRegistry() *Registry {
	return &Registry{
		executors: make(map[string]ActionExecutor),
	}
}

func (r *Registry) Register(name string, executor ActionExecutor) {
	r.executors[name] = executor
}

func (r *Registry) Get(name string) (ActionExecutor, error) {
	exec, exists := r.executors[name]
	if !exists {
		return nil, fmt.Errorf("Unknown action type: %s", name)
	}
	return exec, nil
}

func (r *Registry) Count() int {
	return len(r.executors)
}

func (r *Registry) Types() []string {
	types := make([]string, 0, len(r.executors))
	for k := range r.executors {
		types = append(types, k)
	}
	sort.Strings(types)
	return types
}
