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
	Results  map[string]string
	queueLen map[string]int
}

// NewState 建一份状态。
func NewState() *State {
	return &State{
		Tenants:  map[string]Tenant{},
		InUse:    map[string]int{},
		Results:  map[string]string{},
		queueLen: map[string]int{},
	}
}

// AddTenant 登记一个租户。
func (st *State) AddTenant(name string, limit, queueCap int) {
	st.Tenants[name] = Tenant{Limit: limit, QueueCap: queueCap}
}

// Acquire 申请一个槽位。
func (st *State) Acquire(tenant, reqID string, c *Counter) string {
	// 幂等：同一请求 ID 只认第一次的结果。
	c.Scanned++
	if result, ok := st.Results[reqID]; ok {
		return result
	}

	switch {
	case st.InUse[tenant] < st.Tenants[tenant].Limit:
		st.InUse[tenant]++
		st.Results[reqID] = "admitted"
		return "admitted"
	case st.queueLen[tenant] < st.Tenants[tenant].QueueCap:
		st.Seq++
		st.Pending = append(st.Pending, Pending{
			Ticket: fmt.Sprintf("q%d", st.Seq),
			Tenant: tenant,
			ReqID:  reqID,
		})
		st.queueLen[tenant]++
		st.Results[reqID] = "queued"
		return "queued"
	default:
		st.Results[reqID] = "rejected"
		return "rejected"
	}
}

// Release 释放一个槽位并放行等待者，返回被放行的请求 ID。
func (st *State) Release(tenant string, c *Counter) string {
	if st.InUse[tenant] > 0 {
		st.InUse[tenant]--
	}

	idx := -1
	for i := range st.Pending {
		c.Scanned++
		if st.InUse[st.Pending[i].Tenant] < st.Tenants[st.Pending[i].Tenant].Limit {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ""
	}
	req := st.Pending[idx]
	st.Pending = append(st.Pending[:idx], st.Pending[idx+1:]...)
	st.queueLen[req.Tenant]--
	st.InUse[req.Tenant]++
	return req.ReqID
}

// InUseOf 返回某租户的在用并发数。
func (st *State) InUseOf(tenant string, c *Counter) int {
	c.Scanned++
	return st.InUse[tenant]
}
