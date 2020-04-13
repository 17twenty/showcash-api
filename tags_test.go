package showcash

import (
	"reflect"
	"testing"
)

func Test_getTags(t *testing.T) {
	type args struct {
		tags []string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "The 'People Suck' edition",
			args: args{tags: []string{"hello", "cunt", "piss", "nigger", "piss balls", "fuck", "world"}},
			want: []string{"hello", "piss", "fuck", "world"},
		}, {
			name: "Porn URL check",
			args: args{tags: []string{"pinkspornlist.com", "sexy"}},
			want: []string{"sexy"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanTags(tt.args.tags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cleanTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
