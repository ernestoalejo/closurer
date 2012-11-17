
if [[ "$1" = "server" ]]; then
  go test -i && \
  go test -c && \
  ./closurer.test -test.memprofile mem.prof -test.cpuprofile cpu.prof -gocheck.b -gocheck.f ServeSuite --conf client/config.xml --targets dev
  exit
fi

echo "Target not recgonized"

