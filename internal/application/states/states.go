package states

const (
	StateHome = iota
	StateStart
	StateAddCategory
	StateRemoveCategory
	StateChoosingCategoryForHelp
	StateFormingRequestForHelp
	StateConfirmationRequestForHelp
	StateSendingRequestForHelp
)
