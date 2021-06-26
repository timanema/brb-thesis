# Message efficient Byzantine Reliable Broadcast protocols on known topologies

## Usage
The program can be used in different ways. They will all be briefly discussed in this section.

### Using the provided binaries
We provide the following pre-compiled binaries:
```bash
rp-runner: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, Go BuildID=wvtAbWF_VnSb72SWzwwe/Ih6Z_Rd_PXnEqU-S67SD/_80zGsqfuKHsM5Te-4z8/8il9Jvp0uIkytvXpg_rf, not stripped
rp-runner.exe: PE32+ executable (console) x86-64 (stripped to external PDB), for MS Windows
rp-runner-mac: Mach-O 64-bit x86_64 executable
rp-runner-mac-m1: Mach-O 64-bit arm64 executable, flags:<|DYLDLINK|PIE>
```
Note that these binaries are all compiled on the same 64-bit Linux system, so naturally only the regular `rp-runner` binary is tested. 
The cross-compiled binaries are untested, but no issues are expected.

The binary provides a single command which can either run a single configuration, or one of the predefined suites used for evaluation
in the paper. The help page of the command provides all functionality:
```bash
NAME:
   rp-runner run - Run a test

USAGE:
   rp-runner run [command options] [arguments...]

OPTIONS:
   --template value                select the template to use: brachaDolevIndividualTests | brachaDolevFullTests | brachaDolevScaleTests | dolevIndividualTests | dolevFullTests | dolevScaleTests | brachaIndividualTests | brachaFullTests | brachaScaleTests
   --protocol value, -p value      select the template to use: dolev | bracha | brachaDolev (default: dolev) (default: dolev)
   --generator value, --gen value  select the template to use: randomRegular | multiPartite | fullyConnected | generalizedWheel (default: randomRegular) (default: randomRegular)
   --skip value                    set the amount of template tests to skip (default: 0)
   --runs value                    set the amount of times to run tests (default: 5)
   --nodes value, -n value         amount of nodes (default: 25)
   --connectivity value, -k value  network connectivity (default: 8)
   --degree value, --deg value     network connectivity (degree) (default: k)
   --byzantine value, -f value     amount of byzantine nodes (default: 3)
   --payload value, --ps value     payload size (in bytes) (default: 12)
   --verbosity value, -v value     set verbosity (0, 1, 2, 3) (default: 1)
   --multiple                      enable the use of multiple (N-F) transmitters (default: false)
   --cache                         use graph cache (default: false)
   --ord1                          enable ord1 (filtering of subpaths) (default: false)
   --ord2                          enable ord2 (single hop to neighbours) (default: false)
   --ord3                          enable ord3 (next hop merge) (default: false)
   --ord4                          enable ord4 (path reuse) (default: false)
   --ord5                          enable ord5 (relay merging) (default: false)
   --ord6                          enable ord6 (payload merging) (default: false)
   --ord7                          enable ord7 (implicit paths) (default: false)
   --orb1                          enable orb1 (implicit echo) (default: false)
   --orb2                          enable orb2 (minimal subset) (default: false)
   --orbd1                         enable orbd1 (partial broadcast) (default: false)
   --orbd2                         enable orbd2 (bracha dolev merge) (default: false)
   --no-color                      disable color printing to console (default: false)
   --help, -h                      show help (default: false)
```

By default all optimizations are disabled, but every single one can be enabled independently. The CLI program has not been extensively tested,
so it will not warn of invalid combinations or parameters. 

Note that the template suites may require a large amount of memory (>16GiB), which is the reason some suites were ran in separate runs
on the evaluation system. This can be done by running as much of the suite is possible until no more system resources
are available. Then use the --skip flag to skip the tests that have been completed successfully. 

Note that the generalized wheel is incomplete, as Byzantine node placement is still random. This generator needs to be changed
so that in only places Byzantine nodes at the center.

### Compiling the binary
In case a binary is not working for your platform or you prefer not to run pre-compiled unknown binaries, it is possible to
compile the binary on your own computer. The program is a regular Go program with no dependency on cgo, so it can be (cross-)compiled
with:

```bash
$ CGO_ENABLED=0 go build .
```

### Running from IDE
For more flexibility (and the ability to view code and debug easily) the program can also be started in an IDE. Any IDE 
that supports Go will work, although GoLand from Jetbrains has been used for development and is recommended. Alternatively,
the Go plugin in Intellij (also from Jetbrains) also works. The program is a simple Go program, and has the regular entrypoint
of `main` in `main.go`. The program will by default enter the CLI program, which is undesirable in an IDE. This can be bypassed
by setting the environment variable `MANUAL_RUNNER` to `true`, which has already been done in the provided run configuration
for Jetbrains products.

The `RunnerMain` function (also in `main.go`) will be then be called instead, and this function can be used to experiment
and start the journey in the program. Comments in this function will explain how everything can be modified.