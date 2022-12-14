package models

import (
	"os"

	"github.com/amido/mrbuild/internal/config"
	"github.com/amido/mrbuild/internal/constants"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"

	"github.com/gammazero/workerpool"
)

// App contains objects that the application needs to work
// properly, such as logging
type App struct {
	Logger  *log.Logger
	Workers *workerpool.WorkerPool
}

// ConfigureLogging sets up the logging for the application
// When running in Debug mode, it will be configured to add the caller to the message
func (app *App) ConfigureLogging(logging config.Log) {

	// Initialise a new logger
	app.Logger = log.New()

	// Set the default output to the screen
	app.Logger.Out = os.Stdout

	// Attempt to set the logging level
	ll, err := log.ParseLevel(logging.Level)
	if err != nil {
		ll = log.DebugLevel
		app.HandleError(err, "fatal", "Unable to set logging level")
	}

	// if the log level is set to debug, add the caller to the messages
	if ll == log.TraceLevel {
		app.Logger.SetReportCaller(true)
	}

	// set the logging level
	app.Logger.SetLevel(ll)

	// set the format of the log messages
	switch logging.Format {
	case "json":
		app.Logger.SetFormatter(&log.JSONFormatter{
			TimestampFormat: constants.LoggingTimestamp,
		})
	default:
		app.Logger.SetFormatter(&log.TextFormatter{
			ForceColors:     logging.Colour,
			FullTimestamp:   false,
			TimestampFormat: constants.LoggingTimestamp,
		})

		app.Logger.SetOutput(colorable.NewColorableStdout())
	}

	// if a log file has been set redirect all logs to the file
	if logging.File != "" {
		file, err := os.OpenFile(logging.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			// app.Logger.Out = file
			app.Logger.SetOutput(file)
		} else {
			app.Logger.Warnf("Failed to log to file, defaulting to screen: %s", err.Error())
		}
	}
}

// HandleError handles errors from the application
// This is to ensure that all errors are handled in the same way
func (app *App) HandleError(err error, errorType string, msg ...string) {

	// if no messages have been add the default message
	if len(msg) == 0 {
		msg = append(msg, constants.DefaultErrorMessage)
	}

	var fields map[string]interface{}
	app.HandleErrorWithFields(err, errorType, msg[0], fields)
}

func (app *App) HandleErrorWithFields(err error, errorType string, msg string, fields map[string]interface{}) {

	var message string

	if err != nil {

		if fields == nil {
			fields = make(map[string]interface{})
		}

		// set the message
		if msg != "" {
			message = msg
		} else {
			message = constants.DefaultErrorMessage
		}

		// set the fields that need to be set in the message
		if errorType == "error" ||
			errorType == "fatal" {

			fields["error"] = err
		}

		switch errorType {
		case "error":
			app.Logger.WithFields(fields).Error(message)
		case "fatal":
			app.Logger.WithFields(fields).Fatal(message)
		case "warn":
			app.Logger.WithFields(fields).Warn(message)
		}
	}
}

// ConfigureWorkers sets up the worker pool with the specified number of workers
func (app *App) ConfigureWorkers(count int) {
	app.Logger.WithFields(
		log.Fields{
			"count": count,
		},
	).Debug("Configuring workers in pool")

	app.Workers = workerpool.New(count)
}
