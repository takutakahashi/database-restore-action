package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				filename: "../../test.tar.gz",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := os.Create(tt.args.filename)
			if err != nil {
				t.Error(err)
				return
			}
			r, err := os.Open(fmt.Sprintf("../../misc/%s", filepath.Base(tt.args.filename)))
			if err != nil {
				t.Error(err)
				return
			}
			if _, err := io.Copy(w, r); err != nil {
				t.Error(err)
				return
			}
			w.Close()
			r.Close()
			got, err := extract(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if b, err := os.ReadFile(got); err != nil {
				t.Error(err)
			} else if string(b) != "test\n" {
				t.Errorf("%s", b)
			}
		})
	}
}
