package command_factory

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner"
	"github.com/cloudfoundry-incubator/lattice/ltc/droplet_runner/command_factory/cf_ignore"
	"github.com/cloudfoundry-incubator/lattice/ltc/exit_handler/exit_codes"
	"github.com/cloudfoundry-incubator/lattice/ltc/task_examiner"
	"github.com/cloudfoundry-incubator/lattice/ltc/terminal/colors"
	"github.com/codegangsta/cli"
	"github.com/pivotal-golang/bytefmt"

	"net/url"

	app_runner_command_factory "github.com/cloudfoundry-incubator/lattice/ltc/app_runner/command_factory"
)

var knownBuildpacks map[string]string

func init() {
	knownBuildpacks = map[string]string{
		"go":         "https://github.com/cloudfoundry/go-buildpack.git",
		"java":       "https://github.com/cloudfoundry/java-buildpack.git",
		"python":     "https://github.com/cloudfoundry/python-buildpack.git",
		"ruby":       "https://github.com/cloudfoundry/ruby-buildpack.git",
		"nodejs":     "https://github.com/cloudfoundry/nodejs-buildpack.git",
		"php":        "https://github.com/cloudfoundry/php-buildpack.git",
		"binary":     "https://github.com/cloudfoundry/binary-buildpack.git",
		"staticfile": "https://github.com/cloudfoundry/staticfile-buildpack.git",
	}
}

type DropletRunnerCommandFactory struct {
	app_runner_command_factory.AppRunnerCommandFactory

	taskExaminer  task_examiner.TaskExaminer
	dropletRunner droplet_runner.DropletRunner
	cfIgnore      cf_ignore.CFIgnore
}

type dropletSliceSortedByCreated []droplet_runner.Droplet

func (ds dropletSliceSortedByCreated) Len() int { return len(ds) }
func (ds dropletSliceSortedByCreated) Less(i, j int) bool {
	if ds[j].Created.IsZero() {
		return false
	} else if ds[i].Created.IsZero() {
		return true
	} else {
		return ds[j].Created.Before(ds[i].Created)
	}
}
func (ds dropletSliceSortedByCreated) Swap(i, j int) { ds[i], ds[j] = ds[j], ds[i] }

func NewDropletRunnerCommandFactory(appRunnerCommandFactory app_runner_command_factory.AppRunnerCommandFactory, taskExaminer task_examiner.TaskExaminer, dropletRunner droplet_runner.DropletRunner, cfIgnore cf_ignore.CFIgnore) *DropletRunnerCommandFactory {
	return &DropletRunnerCommandFactory{
		AppRunnerCommandFactory: appRunnerCommandFactory,
		taskExaminer:            taskExaminer,
		dropletRunner:           dropletRunner,
		cfIgnore:                cfIgnore,
	}
}

func (factory *DropletRunnerCommandFactory) MakeListDropletsCommand() cli.Command {
	var listDropletsCommand = cli.Command{
		Name:        "list-droplets",
		Aliases:     []string{"lid", "lsd"},
		Usage:       "List the droplets available to launch",
		Description: "ltc list-droplets",
		Action:      factory.listDroplets,
	}

	return listDropletsCommand
}

func (factory *DropletRunnerCommandFactory) MakeBuildDropletCommand() cli.Command {
	var launchFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "path, p",
			Usage: "Path to droplet source",
			Value: ".",
		},
		cli.StringSliceFlag{
			Name:  "env, e",
			Usage: "Environment variables (can be passed multiple times)",
			Value: &cli.StringSlice{},
		},
		cli.DurationFlag{
			Name:  "timeout, t",
			Usage: "Polling timeout for app to start",
			Value: app_runner_command_factory.DefaultPollingTimeout,
		},
	}

	var buildDropletCommand = cli.Command{
		Name:        "build-droplet",
		Aliases:     []string{"bd"},
		Usage:       "Build droplet",
		Description: "ltc build-droplet DROPLET_NAME http://buildpack/uri",
		Action:      factory.buildDroplet,
		Flags:       launchFlags,
	}

	return buildDropletCommand
}

func (factory *DropletRunnerCommandFactory) MakeLaunchDropletCommand() cli.Command {
	var launchFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "working-dir, w",
			Usage: "Working directory for container",
			Value: "/home/vcap/app",
		},
		cli.BoolFlag{
			Name:  "run-as-root, r",
			Usage: "Runs in the context of the root user",
		},
		cli.StringSliceFlag{
			Name:  "env, e",
			Usage: "Environment variables (can be passed multiple times)",
			Value: &cli.StringSlice{},
		},
		cli.IntFlag{
			Name:  "cpu-weight, c",
			Usage: "Relative CPU weight for the container (valid values: 1-100)",
			Value: 100,
		},
		cli.IntFlag{
			Name:  "memory-mb, m",
			Usage: "Memory limit for container in MB",
			Value: 128,
		},
		cli.IntFlag{
			Name:  "disk-mb, d",
			Usage: "Disk limit for container in MB",
			Value: 0,
		},
		cli.StringFlag{
			Name:  "ports, p",
			Usage: "Ports to expose on the container (comma delimited)",
		},
		cli.IntFlag{
			Name:  "monitor-port, M",
			Usage: "Selects the port used to healthcheck the app",
		},
		cli.StringFlag{
			Name: "monitor-url, U",
			Usage: "Uses HTTP to healthcheck the app\n\t\t" +
				"format is: port:/path/to/endpoint",
		},
		cli.DurationFlag{
			Name:  "monitor-timeout",
			Usage: "Timeout for the app healthcheck",
			Value: time.Second,
		},
		cli.StringFlag{
			Name: "routes, R",
			Usage: "Route mappings to exposed ports as follows:\n\t\t" +
				"--routes=80:web,8080:api will route web to 80 and api to 8080",
		},
		cli.IntFlag{
			Name:  "instances, i",
			Usage: "Number of application instances to spawn on launch",
			Value: 1,
		},
		cli.BoolFlag{
			Name:  "no-monitor",
			Usage: "Disables healthchecking for the app",
		},
		cli.BoolFlag{
			Name:  "no-routes",
			Usage: "Registers no routes for the app",
		},
		cli.DurationFlag{
			Name:  "timeout, t",
			Usage: "Polling timeout for app to start",
			Value: app_runner_command_factory.DefaultPollingTimeout,
		},
	}

	var buildDropletCommand = cli.Command{
		Name:        "launch-droplet",
		Aliases:     []string{"ld"},
		Usage:       "Launch droplet",
		Description: "ltc launch-droplet APP_NAME DROPLET_NAME",
		Action:      factory.launchDroplet,
		Flags:       launchFlags,
	}

	return buildDropletCommand
}

func (factory *DropletRunnerCommandFactory) MakeRemoveDropletCommand() cli.Command {
	var removeDropletCommand = cli.Command{
		Name:        "remove-droplet",
		Aliases:     []string{"rd", "rmd"},
		Usage:       "Remove droplet",
		Description: "ltc remove-droplet DROPLET_NAME",
		Action:      factory.removeDroplet,
	}

	return removeDropletCommand
}

func (factory *DropletRunnerCommandFactory) listDroplets(context *cli.Context) {
	droplets, err := factory.dropletRunner.ListDroplets()
	if err != nil {
		factory.UI.Say(fmt.Sprintf("Error listing droplets: %s", err))
		factory.ExitHandler.Exit(exit_codes.CommandFailed)
		return
	}

	sort.Sort(dropletSliceSortedByCreated(droplets))

	w := &tabwriter.Writer{}
	w.Init(factory.UI, 12, 8, 1, '\t', 0)

	fmt.Fprintln(w, "Droplet\tCreated At\tSize")
	for _, droplet := range droplets {
		size := bytefmt.ByteSize(uint64(droplet.Size))
		if !droplet.Created.IsZero() {
			fmt.Fprintf(w, "%s\t%s\t%s\n", droplet.Name, droplet.Created.Format("01/02 15:04:05.00"), size)
		} else {
			fmt.Fprintf(w, "%s\t\t%s\n", droplet.Name, size)
		}
	}

	w.Flush()
}

func (factory *DropletRunnerCommandFactory) buildDroplet(context *cli.Context) {
	pathFlag := context.String("path")
	envFlag := context.StringSlice("env")
	timeoutFlag := context.Duration("timeout")
	dropletName := context.Args().First()
	buildpack := context.Args().Get(1)

	if dropletName == "" || buildpack == "" {
		factory.UI.SayIncorrectUsage("")
		factory.ExitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	var buildpackUrl string
	if knownBuildpackUrl, ok := knownBuildpacks[buildpack]; ok {
		buildpackUrl = knownBuildpackUrl
	} else if _, err := url.ParseRequestURI(buildpack); err == nil {
		buildpackUrl = buildpack
	} else {
		factory.UI.SayIncorrectUsage(fmt.Sprintf("invalid buildpack %s", buildpack))
		factory.ExitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	archivePath, err := factory.makeTar(pathFlag)
	if err != nil {
		factory.UI.Say(fmt.Sprintf("Error tarring %s: %s", pathFlag, err))
		factory.ExitHandler.Exit(exit_codes.FileSystemError)
		return
	}

	if err = factory.dropletRunner.UploadBits(dropletName, archivePath); err != nil {
		factory.UI.Say(fmt.Sprintf("Error uploading to %s: %s", dropletName, err))
		factory.ExitHandler.Exit(exit_codes.CommandFailed)
		return
	}

	environment := factory.AppRunnerCommandFactory.BuildEnvironment(envFlag)

	taskName := "build-droplet-" + dropletName
	if err = factory.dropletRunner.BuildDroplet(taskName, dropletName, buildpackUrl, environment); err != nil {
		factory.UI.Say(fmt.Sprintf("Error submitting build of %s: %s", dropletName, err))
		factory.ExitHandler.Exit(exit_codes.CommandFailed)
		return
	}

	factory.UI.SayLine("Submitted build of " + dropletName)

	go factory.TailedLogsOutputter.OutputTailedLogs(taskName)
	defer factory.TailedLogsOutputter.StopOutputting()

	ok, taskState := factory.waitForBuildTask(timeoutFlag, taskName)
	if ok {
		if taskState.Failed {
			factory.UI.SayLine("Build failed: " + taskState.FailureReason)
			factory.ExitHandler.Exit(exit_codes.CommandFailed)
		} else {
			factory.UI.SayLine("Build completed")
		}
	} else {
		factory.UI.Say(colors.Red("Timed out waiting for the build to complete."))
		factory.UI.SayNewLine()
		factory.UI.SayLine("Lattice is still building your application in the background.")

		factory.UI.SayLine(fmt.Sprintf("To view logs:\n\tltc logs %s", taskName))
		factory.UI.SayLine(fmt.Sprintf("To view status:\n\tltc status %s", taskName))
		factory.UI.SayNewLine()
	}
}

func (factory *DropletRunnerCommandFactory) waitForBuildTask(pollTimeout time.Duration, taskName string) (bool, task_examiner.TaskInfo) {
	var (
		taskInfo task_examiner.TaskInfo
		err      error
	)

	ok := factory.pollUntilSuccess(pollTimeout, func() bool {
		taskInfo, err = factory.taskExaminer.TaskStatus(taskName)
		if err != nil {
			factory.UI.Say(colors.Red("Error requesting task status: " + err.Error()))
			return true
		}

		return taskInfo.State != "RUNNING" && taskInfo.State != "PENDING"
	})

	return ok, taskInfo
}

func (factory *DropletRunnerCommandFactory) pollUntilSuccess(pollTimeout time.Duration, pollingFunc func() bool) (ok bool) {
	startingTime := factory.Clock.Now()
	for startingTime.Add(pollTimeout).After(factory.Clock.Now()) {
		if result := pollingFunc(); result {
			return true
		}

		factory.Clock.Sleep(1 * time.Second)
	}
	return false
}

func (factory *DropletRunnerCommandFactory) launchDroplet(context *cli.Context) {
	workingDirFlag := context.String("working-dir")
	envVarsFlag := context.StringSlice("env")
	instancesFlag := context.Int("instances")
	cpuWeightFlag := uint(context.Int("cpu-weight"))
	memoryMBFlag := context.Int("memory-mb")
	diskMBFlag := context.Int("disk-mb")
	portsFlag := context.String("ports")
	runAsRootFlag := context.Bool("run-as-root")
	noMonitorFlag := context.Bool("no-monitor")
	portMonitorFlag := context.Int("monitor-port")
	urlMonitorFlag := context.String("monitor-url")
	monitorTimeoutFlag := context.Duration("monitor-timeout")
	routesFlag := context.String("routes")
	noRoutesFlag := context.Bool("no-routes")
	timeoutFlag := context.Duration("timeout")
	appName := context.Args().Get(0)
	dropletName := context.Args().Get(1)
	terminator := context.Args().Get(2)
	startCommand := context.Args().Get(3)

	var startArgs []string

	switch {
	case len(context.Args()) < 2:
		factory.UI.SayIncorrectUsage("APP_NAME and DROPLET_NAME are required")
		factory.ExitHandler.Exit(exit_codes.InvalidSyntax)
		return
	case startCommand != "" && terminator != "--":
		factory.UI.SayIncorrectUsage("'--' Required before start command")
		factory.ExitHandler.Exit(exit_codes.InvalidSyntax)
		return
	case len(context.Args()) > 4:
		startArgs = context.Args()[4:]
	case cpuWeightFlag < 1 || cpuWeightFlag > 100:
		factory.UI.SayIncorrectUsage("Invalid CPU Weight")
		factory.ExitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	exposedPorts, err := factory.parsePortsFromArgs(portsFlag)
	if err != nil {
		factory.UI.Say(err.Error())
		factory.ExitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	monitorConfig, err := factory.GetMonitorConfig(exposedPorts, portMonitorFlag, noMonitorFlag, urlMonitorFlag, monitorTimeoutFlag)
	if err != nil {
		factory.UI.Say(err.Error())
		factory.ExitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	routeOverrides, err := factory.ParseRouteOverrides(routesFlag)
	if err != nil {
		factory.UI.Say(err.Error())
		factory.ExitHandler.Exit(exit_codes.InvalidSyntax)
		return
	}

	appEnvironmentParams := app_runner.AppEnvironmentParams{
		EnvironmentVariables: factory.BuildAppEnvironment(envVarsFlag, appName),
		Privileged:           runAsRootFlag,
		Monitor:              monitorConfig,
		Instances:            instancesFlag,
		CPUWeight:            cpuWeightFlag,
		MemoryMB:             memoryMBFlag,
		DiskMB:               diskMBFlag,
		ExposedPorts:         exposedPorts,
		WorkingDir:           workingDirFlag,
		RouteOverrides:       routeOverrides,
		NoRoutes:             noRoutesFlag,
	}

	if err := factory.dropletRunner.LaunchDroplet(appName, dropletName, startCommand, startArgs, appEnvironmentParams); err != nil {
		factory.UI.Say(fmt.Sprintf("Error launching app %s from droplet %s: %s", appName, dropletName, err))
		factory.ExitHandler.Exit(exit_codes.CommandFailed)
		return
	}

	factory.WaitForAppCreation(appName, timeoutFlag, instancesFlag, noRoutesFlag, routeOverrides)
}

func (factory *DropletRunnerCommandFactory) removeDroplet(context *cli.Context) {
	dropletName := context.Args().First()

	err := factory.dropletRunner.RemoveDroplet(dropletName)
	if err != nil {
		factory.UI.Say(fmt.Sprintf("Error removing droplet %s: %s", dropletName, err))
		factory.ExitHandler.Exit(exit_codes.CommandFailed)
		return
	}

	factory.UI.SayLine("Droplet removed")
}

func (factory *DropletRunnerCommandFactory) makeTar(contentsPath string) (string, error) {
	tmpPath, err := ioutil.TempDir(os.TempDir(), "build-bits")
	if err != nil {
		return "", err
	}

	fileWriter, err := os.OpenFile(filepath.Join(tmpPath, "build-bits.tar"), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	tarWriter := tar.NewWriter(fileWriter)
	defer tarWriter.Close()

	contentsFileInfo, err := os.Stat(contentsPath)
	if err != nil {
		return "", err
	}

	if contentsFileInfo.IsDir() {
		if ignoreFile, err := os.Open(filepath.Join(contentsPath, ".cfignore")); err == nil {
			if err := factory.cfIgnore.Parse(ignoreFile); err != nil {
				return "", err
			}
		}

		err = filepath.Walk(contentsPath, func(fullPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relativePath, err := filepath.Rel(contentsPath, fullPath)
			if err != nil {
				return err
			}

			if factory.cfIgnore.ShouldIgnore(relativePath) {
				return nil
			}

			if err := addFileToTar(fileWriter, tarWriter, info, relativePath, fullPath); err != nil {
				return err
			}

			return nil
		})
	} else {
		err = addFileToTar(fileWriter, tarWriter, contentsFileInfo, path.Base(contentsPath), contentsPath)
	}

	return fileWriter.Name(), err
}

func addFileToTar(fileWriter *os.File, tarWriter *tar.Writer, info os.FileInfo, relativePath, fullPath string) error {
	if relativePath == fileWriter.Name() || relativePath == "." || relativePath == ".." {
		return nil
	}

	if h, err := tar.FileInfoHeader(info, fullPath); err == nil {
		h.Name = relativePath
		if err := tarWriter.WriteHeader(h); err != nil {
			return err
		}
	}

	if info.IsDir() {
		return nil
	}

	li, err := os.Lstat(fullPath)
	if err != nil {
		return err
	}
	if li.Mode()&os.ModeSymlink == os.ModeSymlink {
		return nil
	}

	fr, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer fr.Close()
	if _, err := io.Copy(tarWriter, fr); err != nil {
		return err
	}

	return nil
}

func (factory *DropletRunnerCommandFactory) parsePortsFromArgs(portsFlag string) ([]uint16, error) {
	if portsFlag != "" {
		portStrings := strings.Split(portsFlag, ",")
		sort.Strings(portStrings)

		convertedPorts := []uint16{}
		for _, p := range portStrings {
			intPort, err := strconv.Atoi(p)
			if err != nil || intPort > 65535 {
				return []uint16{}, errors.New(app_runner_command_factory.InvalidPortErrorMessage)
			}
			convertedPorts = append(convertedPorts, uint16(intPort))
		}
		return convertedPorts, nil
	}

	factory.UI.Say(fmt.Sprintf("No port specified. Defaulting to 8080.\n"))

	return []uint16{8080}, nil
}
