package bulkhead

import "testing"

func TestAdmitUnderLimit(t *testing.T) {
	st := NewState()
	st.AddTenant("A", 1, 5)
	if got := st.Acquire("A", "r1", &Counter{}); got != "admitted" {
		t.Fatalf("限额内应准入，得到 %q", got)
	}
}

func TestInUseAfterAdmit(t *testing.T) {
	st := NewState()
	st.AddTenant("A", 1, 5)
	c := &Counter{}
	st.Acquire("A", "r1", c)
	if got := st.InUseOf("A", c); got != 1 {
		t.Fatalf("在用并发应是 1，得到 %d", got)
	}
}

func TestQueueWhenFull(t *testing.T) {
	st := NewState()
	st.AddTenant("A", 1, 5)
	c := &Counter{}
	st.Acquire("A", "r1", c)
	if got := st.Acquire("A", "r2", c); got != "queued" {
		t.Fatalf("满额应排队，得到 %q", got)
	}
}

func TestTenantsStored(t *testing.T) {
	st := NewState()
	st.AddTenant("A", 2, 3)
	if st.Tenants["A"].Limit != 2 {
		t.Fatalf("租户限额应记下：%+v", st.Tenants)
	}
}

func TestCounterCounts(t *testing.T) {
	st := NewState()
	st.AddTenant("A", 1, 5)
	c := &Counter{}
	st.Acquire("A", "r1", c)
	if c.Scanned == 0 {
		t.Fatal("应统计比较次数")
	}
}

func TestReleaseEmptyReturnsEmpty(t *testing.T) {
	st := NewState()
	st.AddTenant("A", 1, 5)
	if got := st.Release("A", &Counter{}); got != "" {
		t.Fatalf("无等待者应返回空，得到 %q", got)
	}
}
