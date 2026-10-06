// Package bulkhead 按租户做并发舱壁。
package bulkhead

import "fmt"

// Counter 记录比较次数。
type Counter struct {
	Scanned int
}

// Tenant 是一个租户的并发与队列上限。
type Tenant struct {
	Limit    int
	QueueCap int
}

// Pending 是排队中的请求。
type Pending struct {
	Ticket string
	Tenant string
	ReqID  string
}

// State 是舱壁状态。
type State struct {
	Tenants map[string]Tenant
	InUse   map[string]int
	Pending []Pending
	Seq     int

	queued  map[string]int
	results map[string]string
}

// NewState 建一份状态。
func NewState() *State {
	return &State{
		Tenants: map[string]Tenant{},
		InUse:   map[string]int{},
		queued:  map[string]int{},
		results: map[string]string{},
	}
}

// AddTenant 登记一个租户。
func (st *State) AddTenant(name string, limit, queueCap int) {
	st.Tenants[name] = Tenant{Limit: limit, QueueCap: queueCap}
}

// Acquire 申请一个槽位。
func (st *State) Acquire(tenant, reqID string, c *Counter) string {
	c.Scanned++
	if res, ok := st.results[reqID]; ok {
		return res
	}
	t := st.Tenants[tenant]
	c.Scanned++
	if st.InUse[tenant] < t.Limit {
		st.InUse[tenant]++
		st.results[reqID] = "admitted"
		return "admitted"
	}
	c.Scanned++
	if st.queued[tenant] >= t.QueueCap {
		st.results[reqID] = "rejected"
		return "rejected"
	}
	st.Seq++
	st.Pending = append(st.Pending, Pending{
		Ticket: fmt.Sprintf("q%d", st.Seq),
		Tenant: tenant,
		ReqID:  reqID,
	})
	st.queued[tenant]++
	st.results[reqID] = "queued"
	return "queued"
}

// Release 释放一个槽位并放行等待者，返回被放行的请求 ID。
func (st *State) Release(tenant string, c *Counter) string {
	if st.InUse[tenant] > 0 {
		st.InUse[tenant]--
	}
	for i, p := range st.Pending {
		c.Scanned++
		if st.InUse[p.Tenant] < st.Tenants[p.Tenant].Limit {
			st.InUse[p.Tenant]++
			st.queued[p.Tenant]--
			st.Pending = append(st.Pending[:i], st.Pending[i+1:]...)
			return p.ReqID
		}
	}
	return ""
}

// InUseOf 返回某租户的在用并发数。
func (st *State) InUseOf(tenant string, c *Counter) int {
	c.Scanned++
	return st.InUse[tenant]
}
