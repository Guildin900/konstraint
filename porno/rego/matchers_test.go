package rego

import (
	"reflect"
	"testing"
)

func TestMultilineKindMatchers(t *testing.T) {
	comments := []string{
		"@kinds core/Pod apps/Deployment",
		"@kinds apps/StatefulSet",
	}
	rego := Rego{
		headerComments: comments,
	}

	expected := KindMatchers{
		{APIGroup: "", Kinds: []string{"Pod"}},
		{APIGroup: "apps", Kinds: []string{"Deployment", "StatefulSet"}},
	}

	matchers, err := rego.Matchers()
	if err != nil {
		t.Fatal(err)
	}
	actual := matchers.KindMatchers

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Unexpected KindMatchers. expected %v, actual %v.", expected, actual)
	}
}

func TestKindMatchers(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    KindMatchers
	}{
		{
			"core_Pod",
			"@kinds core/Pod",
			KindMatchers{{APIGroup: "", Kinds: []string{"Pod"}}},
		},
		{
			"core_Pod,core_Pod",
			"@kinds core/Pod core/Pod",
			KindMatchers{{APIGroup: "", Kinds: []string{"Pod"}}},
		},
		{
			"apps_Deployment,apps_StatefulSet",
			"@kinds apps/Deployment apps/StatefulSet",
			KindMatchers{{APIGroup: "apps", Kinds: []string{"Deployment", "StatefulSet"}}},
		},
		{
			"apps_StatefulSet,apps_Deployment",
			"@kinds apps/StatefulSet apps/Deployment",
			KindMatchers{{APIGroup: "apps", Kinds: []string{"Deployment", "StatefulSet"}}},
		},
		{
			"apps_Deployment,apps_StatefulSet,core_Pod",
			"@kinds apps/Deployment apps/StatefulSet core/Pod",
			KindMatchers{
				{APIGroup: "", Kinds: []string{"Pod"}},
				{APIGroup: "apps", Kinds: []string{"Deployment", "StatefulSet"}},
			},
		},
		{
			"apps_Deployment,core_Pod,apps_StatefulSet",
			"@kinds apps/Deployment core/Pod apps/StatefulSet",
			KindMatchers{
				{APIGroup: "", Kinds: []string{"Pod"}},
				{APIGroup: "apps", Kinds: []string{"Deployment", "StatefulSet"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchers, err := Rego{headerComments: []string{tt.comment}}.Matchers()
			if err != nil {
				t.Fatal(err)
			}
			actual := matchers.KindMatchers
			if !reflect.DeepEqual(tt.want, actual) {
				t.Errorf("Unexpected KindMatchers. expected %v, actual %v.", tt.want, actual)
			}
		})
	}
}

func TestGetMatchLabelsMatcher(t *testing.T) {
	comments := []string{
		"@matchlabels team=a app.kubernetes.io/name=test",
		"@matchlabels example.com/env=production",
	}
	rego := Rego{
		headerComments: comments,
	}

	expected := MatchLabelsMatcher{
		"team":                   "a",
		"app.kubernetes.io/name": "test",
		"example.com/env":        "production",
	}

	matchers, err := rego.Matchers()
	if err != nil {
		t.Fatal(err)
	}
	actual := matchers.MatchLabelsMatcher

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Unexpected MatchLabelMatcher. expected %v, actual %v.", expected, actual)
	}
}

func TestGetMatchExpressionsMatcher(t *testing.T) {
	testCases := []struct {
		name      string
		comments  []string
		expected  []MatchExpressionMatcher
		wantError bool
	}{
		{
			name: "Empty",
		},
		{
			name:     "Single",
			comments: []string{"@matchExpression foo In bar,baz"},
			expected: []MatchExpressionMatcher{
				{Key: "foo", Operator: "In", Values: []string{"bar", "baz"}},
			},
		},
		{
			name: "DoubleWithUnrelated",
			comments: []string{
				"@matchExpression foo In bar,baz",
				"@matchExpression doggos Exists",
				"unrelated comment",
			},
			expected: []MatchExpressionMatcher{
				{Key: "foo", Operator: "In", Values: []string{"bar", "baz"}},
				{Key: "doggos", Operator: "Exists"},
			},
		},
		{
			name: "TooFewParams",
			comments: []string{
				"@matchExpression foo",
			},
			wantError: true,
		},
		{
			name: "TooManyParams",
			comments: []string{
				"@matchExpression foo In bar baz",
			},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rego := Rego{headerComments: tc.comments}
			matchers, err := rego.Matchers()
			if (err != nil && !tc.wantError) || (err == nil && tc.wantError) {
				t.Errorf("Unexpected error state, have %v want %v", !tc.wantError, tc.wantError)
			}
			actual := matchers.MatchExpressionsMatcher
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Unexpected MatchExpressionsMatcher, have %v want %v", actual, tc.expected)
			}
		})
	}
}

func TestGetStringListMatcher(t *testing.T) {
	testCases := []struct {
		desc      string
		tag       string
		comment   string
		expected  []string
		wantError bool
	}{
		{
			desc:      "InvalidTag",
			wantError: true, // will error without a match on the tag
		},
		{
			desc:      "NoValuesSupplied",
			tag:       "@foo",
			comment:   "@foo     ",
			wantError: true,
		},
		{
			desc:     "Valid",
			tag:      "@foo",
			comment:  "@foo bar baz",
			expected: []string{"bar", "baz"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			actual, err := getStringListMatcher(tc.tag, tc.comment)
			if (err != nil && !tc.wantError) || (err == nil && tc.wantError) {
				t.Errorf("Unexpected error state, have %v want %v", !tc.wantError, tc.wantError)
			}
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Unexpected values, have %v want %v", actual, tc.expected)
			}
		})
	}
}
