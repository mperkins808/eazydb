package eazydb

import (
	"fmt"
)

type Condition struct {
	clause string
}

func (q *Query) Where(conditions ...Condition) *Query {
	q.conditions = append(q.conditions, conditions...)
	return q
}

func (c *Condition) Or(condition Condition) *Condition {
	if c.clause == "" {
		return c
	}
	c.clause = fmt.Sprintf("(%s OR %s)", c.clause, condition.clause)
	return c
}

type StrCond struct {
	name string
}

func String(name string) *StrCond {
	return &StrCond{
		name: name,
	}
}

func (s *StrCond) Equals(val string) *Condition {
	return &Condition{
		clause: fmt.Sprintf("%s = '%v'", s.name, val),
	}
}

func (s *StrCond) NotEqual(val string) *Condition {
	return &Condition{
		clause: fmt.Sprintf("%s != '%v'", s.name, val),
	}
}

func (s *StrCond) Contains(val string) *Condition {
	search := "%" + val + "%"
	return &Condition{
		clause: fmt.Sprintf("%s LIKE '%v'", s.name, search),
	}
}

func (s *StrCond) StartsWith(val string) *Condition {
	search := val + "%"
	return &Condition{
		clause: fmt.Sprintf("%s LIKE '%v'", s.name, search),
	}
}

func (s *StrCond) EndsWith(val string) *Condition {
	search := "%" + val
	return &Condition{
		clause: fmt.Sprintf("%s LIKE '%v'", s.name, search),
	}
}

type IntCond struct {
	name string
}

func Int(name string) *IntCond {
	return &IntCond{
		name: name,
	}
}

func (i *IntCond) Equals(val int) *Condition {
	return &Condition{
		clause: fmt.Sprintf("%s = %v", i.name, val),
	}
}

func (i *IntCond) GreaterThan(val int) *Condition {
	return &Condition{
		clause: fmt.Sprintf("%s > %v", i.name, val),
	}
}

func (i *IntCond) GreaterThanOrEqual(val int) *Condition {
	return &Condition{
		clause: fmt.Sprintf("%s >= %v", i.name, val),
	}
}

func (i *IntCond) LessThan(val int) *Condition {
	return &Condition{
		clause: fmt.Sprintf("%s < %v", i.name, val),
	}
}

func (i *IntCond) LessThanOrEqual(val int) *Condition {
	return &Condition{
		clause: fmt.Sprintf("%s < %v", i.name, val),
	}
}

func (i *IntCond) NotEqual(val int) *Condition {
	return &Condition{
		clause: fmt.Sprintf("%s != %v", i.name, val),
	}
}
