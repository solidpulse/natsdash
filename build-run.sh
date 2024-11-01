LDFLAGS="-ldflags '-X github.com/solidpulse/natsdash/ds.Version=1.0.0'"
go build -gcflags=all="-N -l" -o natsdash
./natsdash