package bulkhead

import "fmt"

// Sample 是一个固定场景，用来复现题面里的锚点。
type Sample struct {
	Name string
	Run  func() map[string]any
}

// Samples 返回全部场景。
func Samples() []Sample {
	return []Sample{
		{Name: "idem", Run: func() map[string]any {
			st := NewState()
			st.AddTenant("A", 1, 5)
			c := &Counter{}
			st.Acquire("A", "r1", c)
			return map[string]any{"second": st.Acquire("A", "r1", c)}
		}},
		{Name: "queuecap", Run: func() map[string]any {
			st := NewState()
			st.AddTenant("A", 1, 1)
			c := &Counter{}
			st.Acquire("A", "r1", c)
			st.Acquire("A", "r2", c)
			return map[string]any{"third": st.Acquire("A", "r3", c)}
		}},
		{Name: "release", Run: func() map[string]any {
			st := NewState()
			st.AddTenant("A", 1, 5)
			c := &Counter{}
			st.Acquire("A", "r1", c)
			st.Release("A", c)
			return map[string]any{"after": st.Acquire("A", "r2", c)}
		}},
		{Name: "fair", Run: func() map[string]any {
			st := NewState()
			st.AddTenant("A", 1, 5)
			c := &Counter{}
			st.Acquire("A", "r1", c)
			st.Acquire("A", "q1", c)
			st.Acquire("A", "q2", c)
			return map[string]any{"admitted": st.Release("A", c)}
		}},
	}
}

// Work 跑规模线场景。
func Work(n int) map[string]any {
	st := NewState()
	st.AddTenant("A", 100, 100000)
	c := &Counter{}
	for i := 0; i < n; i++ {
		st.Acquire("A", fmt.Sprintf("r%04d", i), c)
	}
	return map[string]any{"requests": n, "scanned": c.Scanned}
}
