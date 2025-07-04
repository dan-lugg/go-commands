package commands

// CommandRes is an interface that represents the result of a command.
// It can be implemented by any type that represents the output of a command.
type CommandRes any

// CommandReq is a generic interface representing a request for a command.
// It is parameterized by TRes, which must implement the CommandRes interface.
type CommandReq[TRes CommandRes] any
