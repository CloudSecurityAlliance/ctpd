package ctp

import (
    "strings"
)

type Tags []string

func NewTags(tags ...string) Tags {
    return Tags(tags)
}

func (t *Tags) Append(tags Tags) {
    *t = append(*t, tags...)
}

func (t Tags) HasWildcard() bool {
    for _, token_tag := range t {
        if token_tag == "*" {
            return true
        }
    }
    return false
}

func (t Tags)String() string {
    return strings.Join([]string(t),",")
}

func MatchTags(src_tags, dst_tags Tags) bool {
    for _, src := range src_tags {
        if src == "*" {
            return true
        }
        for _, dst := range dst_tags {
            if src == dst {
                return true
            }
        }
    }
    return false
}

func (t Tags) WithPrefix(prefix string) []string {
    var result []string

    for _, token_tag := range t {
        if strings.HasPrefix(token_tag, prefix) {
            result = append(result, token_tag)
        }
    }
    return result
}
