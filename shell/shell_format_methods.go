package shell

import "os/exec"

// AUTO-GENERATED WITH static_compose [-in . -out shell_format_methods.go]

// Outp is equivalent to sh.Out(ScriptPrint(vs...))
func (sh *MockShell) Outp(vs ...interface{}) string {
	return sh.Out(ScriptPrint(vs...))
}

// Outf is equivalent to sh.Out(ScriptPrintf(scriptformat, vs...))
func (sh *MockShell) Outf(scriptformat string, vs ...interface{}) string {
	return sh.Out(ScriptPrintf(scriptformat, vs...))
}

// Outt is equivalent to sh.Out(ScriptTemplate(template, vars))
func (sh *MockShell) Outt(template string, vars Lookuper) string {
	return sh.Out(ScriptTemplate(template, vars))
}

// OutStatusp is equivalent to sh.OutStatus(ScriptPrint(vs...))
func (sh *MockShell) OutStatusp(vs ...interface{}) (string, error) {
	return sh.OutStatus(ScriptPrint(vs...))
}

// OutStatusf is equivalent to sh.OutStatus(ScriptPrintf(scriptformat, vs...))
func (sh *MockShell) OutStatusf(scriptformat string, vs ...interface{}) (string, error) {
	return sh.OutStatus(ScriptPrintf(scriptformat, vs...))
}

// OutStatust is equivalent to sh.OutStatus(ScriptTemplate(template, vars))
func (sh *MockShell) OutStatust(template string, vars Lookuper) (string, error) {
	return sh.OutStatus(ScriptTemplate(template, vars))
}

// OutErrStatusp is equivalent to sh.OutErrStatus(ScriptPrint(vs...))
func (sh *MockShell) OutErrStatusp(vs ...interface{}) (string, string, error) {
	return sh.OutErrStatus(ScriptPrint(vs...))
}

// OutErrStatusf is equivalent to sh.OutErrStatus(ScriptPrintf(scriptformat, vs...))
func (sh *MockShell) OutErrStatusf(scriptformat string, vs ...interface{}) (string, string, error) {
	return sh.OutErrStatus(ScriptPrintf(scriptformat, vs...))
}

// OutErrStatust is equivalent to sh.OutErrStatus(ScriptTemplate(template, vars))
func (sh *MockShell) OutErrStatust(template string, vars Lookuper) (string, string, error) {
	return sh.OutErrStatus(ScriptTemplate(template, vars))
}

// Runp is equivalent to sh.Run(ScriptPrint(vs...))
func (sh *MockShell) Runp(vs ...interface{}) error {
	return sh.Run(ScriptPrint(vs...))
}

// Runf is equivalent to sh.Run(ScriptPrintf(scriptformat, vs...))
func (sh *MockShell) Runf(scriptformat string, vs ...interface{}) error {
	return sh.Run(ScriptPrintf(scriptformat, vs...))
}

// Runt is equivalent to sh.Run(ScriptTemplate(template, vars))
func (sh *MockShell) Runt(template string, vars Lookuper) error {
	return sh.Run(ScriptTemplate(template, vars))
}

// Succeedsp is equivalent to sh.Succeeds(ScriptPrint(vs...))
func (sh *MockShell) Succeedsp(vs ...interface{}) bool {
	return sh.Succeeds(ScriptPrint(vs...))
}

// Succeedsf is equivalent to sh.Succeeds(ScriptPrintf(scriptformat, vs...))
func (sh *MockShell) Succeedsf(scriptformat string, vs ...interface{}) bool {
	return sh.Succeeds(ScriptPrintf(scriptformat, vs...))
}

// Succeedst is equivalent to sh.Succeeds(ScriptTemplate(template, vars))
func (sh *MockShell) Succeedst(template string, vars Lookuper) bool {
	return sh.Succeeds(ScriptTemplate(template, vars))
}

// Cmdp is equivalent to sh.Cmd(ScriptPrint(vs...))
func (sh *Shell) Cmdp(vs ...interface{}) *exec.Cmd {
	return sh.Cmd(ScriptPrint(vs...))
}

// Cmdf is equivalent to sh.Cmd(ScriptPrintf(scriptformat, vs...))
func (sh *Shell) Cmdf(scriptformat string, vs ...interface{}) *exec.Cmd {
	return sh.Cmd(ScriptPrintf(scriptformat, vs...))
}

// Cmdt is equivalent to sh.Cmd(ScriptTemplate(template, vars))
func (sh *Shell) Cmdt(template string, vars Lookuper) *exec.Cmd {
	return sh.Cmd(ScriptTemplate(template, vars))
}

// Outp is equivalent to sh.Out(ScriptPrint(vs...))
func (sh *Shell) Outp(vs ...interface{}) string {
	return sh.Out(ScriptPrint(vs...))
}

// Outf is equivalent to sh.Out(ScriptPrintf(scriptformat, vs...))
func (sh *Shell) Outf(scriptformat string, vs ...interface{}) string {
	return sh.Out(ScriptPrintf(scriptformat, vs...))
}

// Outt is equivalent to sh.Out(ScriptTemplate(template, vars))
func (sh *Shell) Outt(template string, vars Lookuper) string {
	return sh.Out(ScriptTemplate(template, vars))
}

// OutStatusp is equivalent to sh.OutStatus(ScriptPrint(vs...))
func (sh *Shell) OutStatusp(vs ...interface{}) (string, error) {
	return sh.OutStatus(ScriptPrint(vs...))
}

// OutStatusf is equivalent to sh.OutStatus(ScriptPrintf(scriptformat, vs...))
func (sh *Shell) OutStatusf(scriptformat string, vs ...interface{}) (string, error) {
	return sh.OutStatus(ScriptPrintf(scriptformat, vs...))
}

// OutStatust is equivalent to sh.OutStatus(ScriptTemplate(template, vars))
func (sh *Shell) OutStatust(template string, vars Lookuper) (string, error) {
	return sh.OutStatus(ScriptTemplate(template, vars))
}

// OutErrStatusp is equivalent to sh.OutErrStatus(ScriptPrint(vs...))
func (sh *Shell) OutErrStatusp(vs ...interface{}) (string, string, error) {
	return sh.OutErrStatus(ScriptPrint(vs...))
}

// OutErrStatusf is equivalent to sh.OutErrStatus(ScriptPrintf(scriptformat, vs...))
func (sh *Shell) OutErrStatusf(scriptformat string, vs ...interface{}) (string, string, error) {
	return sh.OutErrStatus(ScriptPrintf(scriptformat, vs...))
}

// OutErrStatust is equivalent to sh.OutErrStatus(ScriptTemplate(template, vars))
func (sh *Shell) OutErrStatust(template string, vars Lookuper) (string, string, error) {
	return sh.OutErrStatus(ScriptTemplate(template, vars))
}

// Runp is equivalent to sh.Run(ScriptPrint(vs...))
func (sh *Shell) Runp(vs ...interface{}) error {
	return sh.Run(ScriptPrint(vs...))
}

// Runf is equivalent to sh.Run(ScriptPrintf(scriptformat, vs...))
func (sh *Shell) Runf(scriptformat string, vs ...interface{}) error {
	return sh.Run(ScriptPrintf(scriptformat, vs...))
}

// Runt is equivalent to sh.Run(ScriptTemplate(template, vars))
func (sh *Shell) Runt(template string, vars Lookuper) error {
	return sh.Run(ScriptTemplate(template, vars))
}

// Succeedsp is equivalent to sh.Succeeds(ScriptPrint(vs...))
func (sh *Shell) Succeedsp(vs ...interface{}) bool {
	return sh.Succeeds(ScriptPrint(vs...))
}

// Succeedsf is equivalent to sh.Succeeds(ScriptPrintf(scriptformat, vs...))
func (sh *Shell) Succeedsf(scriptformat string, vs ...interface{}) bool {
	return sh.Succeeds(ScriptPrintf(scriptformat, vs...))
}

// Succeedst is equivalent to sh.Succeeds(ScriptTemplate(template, vars))
func (sh *Shell) Succeedst(template string, vars Lookuper) bool {
	return sh.Succeeds(ScriptTemplate(template, vars))
}
