package storage

import (
	"os"
	"testing"
)

func Test_splitExt(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			args: args{
				filename: "test.txt",
			},
			want:  "test",
			want1: ".txt",
		},
		{
			args: args{
				filename: "test.tar.gz",
			},
			want:  "test.tar",
			want1: ".gz",
		},
		{
			args: args{
				filename: "test.tar.gz.aaa.iii.uuu.eee",
			},
			want:  "test.tar.gz.aaa.iii.uuu",
			want1: ".eee",
		},
		{
			args: args{
				filename: "test",
			},
			want:  "test",
			want1: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := splitExt(tt.args.filename)
			if got != tt.want {
				t.Errorf("splitExt() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("splitExt() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_extract(t *testing.T) {
	type args struct {
		filename string
		r        *os.File
	}
	r, err := os.Open("../../aaa.tar.gz")
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				filename: "aaa.tar.gz",
				r:        r,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := extract(tt.args.filename, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			buf := []byte{}
			if _, err := r.Read(buf); err != nil {
				t.Error(err)
				return
			}
			if string(buf) != "test" {
				t.Errorf("%s", buf)
			}
		})
	}
}
