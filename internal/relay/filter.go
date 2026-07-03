package relay

import (
	"encoding/json"
	"fmt"
)

func (f *Filter) Match(event *Event) bool {
	if len(f.IDs) > 0 {
		if !stringInSlice(event.ID, f.IDs) {
			return false
		}
	}
	if len(f.Authors) > 0 {
		if !stringInSlice(event.Pubkey, f.Authors) {
			return false
		}
	}
	if len(f.Kinds) > 0 {
		if !intInSlice(event.Kind, f.Kinds) {
			return false
		}
	}
	for key, vals := range f.Tags {
		if !eventHasTag(event, key, vals) {
			return false
		}
	}
	if f.Since != nil && event.CreatedAt < *f.Since {
		return false
	}
	if f.Until != nil && event.CreatedAt > *f.Until {
		return false
	}
	return true
}

func eventHasTag(event *Event, key string, values []string) bool {
	for _, tag := range event.Tags {
		if len(tag) > 1 && tag[0] == key {
			for _, v := range values {
				if tag[1] == v {
					return true
				}
			}
		}
	}
	return false
}

func stringInSlice(s string, list []string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func intInSlice(n int, list []int) bool {
	for _, item := range list {
		if item == n {
			return true
		}
	}
	return false
}

func (f *Filter) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	if len(f.IDs) > 0 {
		m["ids"] = f.IDs
	}
	if len(f.Authors) > 0 {
		m["authors"] = f.Authors
	}
	if len(f.Kinds) > 0 {
		m["kinds"] = f.Kinds
	}
	if f.Since != nil {
		m["since"] = *f.Since
	}
	if f.Until != nil {
		m["until"] = *f.Until
	}
	if f.Limit != nil {
		m["limit"] = *f.Limit
	}
	for key, vals := range f.Tags {
		if len(vals) > 0 {
			m["#"+key] = vals
		}
	}
	return json.Marshal(m)
}

func (f *Filter) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	f.Tags = make(map[string][]string)
	for key, val := range raw {
		var err error
		switch {
		case key == "ids":
			err = json.Unmarshal(val, &f.IDs)
		case key == "authors":
			err = json.Unmarshal(val, &f.Authors)
		case key == "kinds":
			err = json.Unmarshal(val, &f.Kinds)
		case key == "since":
			err = json.Unmarshal(val, &f.Since)
		case key == "until":
			err = json.Unmarshal(val, &f.Until)
		case key == "limit":
			err = json.Unmarshal(val, &f.Limit)
		case len(key) > 1 && key[0] == '#':
			tagKey := key[1:]
			var vals []string
			if uerr := json.Unmarshal(val, &vals); uerr == nil {
				f.Tags[tagKey] = vals
			}
		default:
			return fmt.Errorf("unknown filter key: %s", key)
		}
		if err != nil {
			return fmt.Errorf("error decoding %s: %w", key, err)
		}
	}
	return nil
}

func (f *Filter) ToJSON() ([]byte, error) {
	return json.Marshal(f)
}

func FilterFromJSON(data []byte) (*Filter, error) {
	var f Filter
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	if f.Tags == nil {
		f.Tags = make(map[string][]string)
	}
	return &f, nil
}
