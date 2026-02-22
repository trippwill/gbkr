package gbkr

// ErrPacingWait indicates that a pacing wait was interrupted, typically
// because the context was cancelled or timed out.
const ErrPacingWait = Error("pacing wait interrupted")
