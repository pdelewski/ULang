package gui

import graphics "runtime/graphics"

// --- Core Types ---

// GuiContext holds minimal state for immediate-mode GUI
type GuiContext struct {
	// Input state (updated each frame)
	MouseX        int32
	MouseY        int32
	MouseDown     bool
	MouseClicked  bool
	MouseReleased bool

	// Widget interaction tracking
	HotID      int32
	ActiveID   int32
	ReleasedID int32 // ID that was active when mouse was released

	// Layout cursor
	CursorX int32
	CursorY int32
	Spacing int32

	// Current style
	Style GuiStyle
}

// WindowState holds position and drag state for a draggable window
type WindowState struct {
	X           int32
	Y           int32
	Width       int32
	Height      int32
	Dragging    bool
	DragOffsetX int32
	DragOffsetY int32
}

// NewWindowState creates a new window state at the given position
func NewWindowState(x int32, y int32, width int32, height int32) WindowState {
	return WindowState{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

// MenuState holds state for menu bar and dropdowns
type MenuState struct {
	OpenMenuID   int32 // ID of currently open menu (0 = none)
	MenuBarX     int32 // Menu bar position for dropdown alignment
	MenuBarY     int32
	MenuBarH     int32 // Menu bar height
	CurrentMenuX int32 // X position of current menu header
	CurrentMenuW int32 // Width of current menu header
	ClickedOutside bool // Flag to close menu on outside click
}

// NewMenuState creates a new menu state
func NewMenuState() MenuState {
	return MenuState{}
}

// GuiStyle defines colors and dimensions for widgets
type GuiStyle struct {
	// Colors
	BackgroundColor   graphics.Color
	TextColor         graphics.Color
	ButtonColor       graphics.Color
	ButtonHoverColor  graphics.Color
	ButtonActiveColor graphics.Color
	CheckboxColor     graphics.Color
	CheckmarkColor    graphics.Color
	SliderTrackColor  graphics.Color
	SliderKnobColor   graphics.Color
	BorderColor       graphics.Color
	FrameBgColor      graphics.Color // Background for input widgets
	TitleBgColor      graphics.Color // Panel/window title background

	// Dimensions
	FontSize     int32
	Padding      int32
	ButtonHeight int32
	SliderHeight int32
	CheckboxSize int32
	FrameRounding int32 // Visual depth simulation
}

// --- Initialization ---

// DefaultStyle returns an ImGui-like style
func DefaultStyle() GuiStyle {
	return GuiStyle{
		// ImGui Dark theme colors (from actual ImGui)
		BackgroundColor:   graphics.NewColor(15, 15, 15, 240),    // Window body - very dark
		TextColor:         graphics.NewColor(255, 255, 255, 255), // White text
		ButtonColor:       graphics.NewColor(66, 150, 250, 102),  // Button normal
		ButtonHoverColor:  graphics.NewColor(66, 150, 250, 200),  // Button hover
		ButtonActiveColor: graphics.NewColor(15, 135, 250, 255),  // Button pressed
		CheckboxColor:     graphics.NewColor(41, 74, 122, 255),   // Checkbox/frame bg
		CheckmarkColor:    graphics.NewColor(66, 150, 250, 255),  // Checkmark color
		SliderTrackColor:  graphics.NewColor(66, 150, 250, 171),  // Slider filled
		SliderKnobColor:   graphics.NewColor(66, 150, 250, 200),  // Slider grab
		BorderColor:       graphics.NewColor(80, 80, 80, 255),    // Borders
		FrameBgColor:      graphics.NewColor(29, 47, 73, 138),    // Input field bg
		TitleBgColor:      graphics.NewColor(41, 74, 137, 255),   // Title bar - blue!
		FontSize:          1,                                     // 8px like classic ImGui
		Padding:           6,
		ButtonHeight:      20,
		SliderHeight:      18,
		CheckboxSize:      16,
		FrameRounding:     2,
	}
}

// NewContext creates a new GUI context with default style
func NewContext() GuiContext {
	return GuiContext{
		Style: DefaultStyle(),
	}
}

// --- Widget ID Generation ---

// GenID generates a unique ID for a widget based on its label
// Uses simple hash for portability across all backends
func GenID(label string) int32 {
	hash := int32(5381) // djb2 hash starting value
	for i := 0; i < len(label); i++ {
		hash = ((hash << 5) + hash) + int32(label[i])
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// --- Input Handling ---

// UpdateInput must be called once per frame before any GUI widgets
// Returns updated context (goany doesn't support pointers)
func UpdateInput(ctx GuiContext, w graphics.Window) GuiContext {
	// Store previous state
	prevDown := ctx.MouseDown

	// Get current mouse state
	x, y, buttons := graphics.GetMouse(w)
	ctx.MouseX = x
	ctx.MouseY = y
	ctx.MouseDown = (buttons & 1) != 0

	// Detect click/release events
	ctx.MouseClicked = ctx.MouseDown && !prevDown
	ctx.MouseReleased = !ctx.MouseDown && prevDown

	// Save which widget was active when mouse released (for click detection)
	ctx.ReleasedID = 0
	if ctx.MouseReleased {
		ctx.ReleasedID = ctx.ActiveID
		ctx.ActiveID = 0
	}

	// Reset hot ID each frame
	ctx.HotID = 0

	return ctx
}

// --- Text Rendering ---

// drawChar renders a single character from the bitmap font
func drawChar(w graphics.Window, charCode int, x int32, y int32, scale int32, color graphics.Color) {
	if charCode < 32 || charCode > 127 {
		charCode = 32
	}
	offset := (charCode - 32) * 8
	font := getFontData()
	for row := int32(0); row < 8; row++ {
		rowData := font[offset+int(row)]
		for col := int32(0); col < 8; col++ {
			if (rowData & (0x80 >> col)) != 0 {
				// Draw scaled pixel
				for sy := int32(0); sy < scale; sy++ {
					for sx := int32(0); sx < scale; sx++ {
						graphics.DrawPoint(w, x+col*scale+sx, y+row*scale+sy, color)
					}
				}
			}
		}
	}
}

// DrawText renders a string at (x, y) with the given color
// scale: 1 = 8px, 2 = 16px, etc.
func DrawText(w graphics.Window, text string, x int32, y int32, scale int32, color graphics.Color) {
	curX := x
	for i := 0; i < len(text); i++ {
		ch := int(text[i])
		drawChar(w, ch, curX, y, scale, color)
		curX = curX + (8 * scale)
	}
}

// TextWidth returns the width in pixels of a text string
func TextWidth(text string, scale int32) int32 {
	return int32(len(text)) * 8 * scale
}

// TextHeight returns the height in pixels (fixed for bitmap font)
func TextHeight(scale int32) int32 {
	return 8 * scale
}

// --- Helper Functions ---

// pointInRect checks if a point is inside a rectangle
func pointInRect(px int32, py int32, x int32, y int32, w int32, h int32) bool {
	return px >= x && px < x+w && py >= y && py < y+h
}

// --- Widgets ---

// Label draws static text
func Label(ctx GuiContext, w graphics.Window, text string, x int32, y int32) {
	DrawText(w, text, x, y, ctx.Style.FontSize, ctx.Style.TextColor)
}

// Button returns updated context and true if the button was clicked this frame
func Button(ctx GuiContext, w graphics.Window, label string, x int32, y int32, width int32, height int32) (GuiContext, bool) {
	id := GenID(label)

	// Hit test
	hovered := pointInRect(ctx.MouseX, ctx.MouseY, x, y, width, height)

	// Update hot/active state
	if hovered {
		ctx.HotID = id
		if ctx.MouseClicked {
			ctx.ActiveID = id
		}
	}

	// Determine visual state and offset for press effect
	var bgColor graphics.Color
	pressOffset := int32(0)
	if ctx.ActiveID == id && hovered {
		bgColor = ctx.Style.ButtonActiveColor
		pressOffset = 1
	} else if ctx.HotID == id {
		bgColor = ctx.Style.ButtonHoverColor
	} else {
		bgColor = ctx.Style.ButtonColor
	}

	// Draw button with ImGui-like frame
	// Dark edge on top-left for depth
	graphics.DrawLine(w, x, y+height-1, x, y, graphics.NewColor(0, 0, 0, 80))
	graphics.DrawLine(w, x, y, x+width-1, y, graphics.NewColor(0, 0, 0, 80))

	// Main button fill
	graphics.FillRect(w, graphics.NewRect(x+1, y+1, width-2, height-2), bgColor)

	// Light edge on bottom-right for depth (subtle)
	graphics.DrawLine(w, x+1, y+height-1, x+width-1, y+height-1, graphics.NewColor(255, 255, 255, 30))
	graphics.DrawLine(w, x+width-1, y+1, x+width-1, y+height-1, graphics.NewColor(255, 255, 255, 30))

	// Draw centered label with press offset
	textW := TextWidth(label, ctx.Style.FontSize)
	textH := TextHeight(ctx.Style.FontSize)
	textX := x + (width-textW)/2 + pressOffset
	textY := y + (height-textH)/2 + pressOffset
	DrawText(w, label, textX, textY, ctx.Style.FontSize, ctx.Style.TextColor)

	// Return context and true if clicked (was active when mouse released while hovered)
	clicked := ctx.ReleasedID == id && ctx.MouseReleased && hovered
	return ctx, clicked
}

// Checkbox draws a checkbox and returns updated context and new value
func Checkbox(ctx GuiContext, w graphics.Window, label string, x int32, y int32, value bool) (GuiContext, bool) {
	id := GenID(label)
	boxSize := ctx.Style.CheckboxSize

	// Hit test (checkbox + label area)
	labelW := TextWidth(label, ctx.Style.FontSize)
	totalW := boxSize + ctx.Style.Padding + labelW
	hovered := pointInRect(ctx.MouseX, ctx.MouseY, x, y, totalW, boxSize)

	// Update hot/active state
	if hovered {
		ctx.HotID = id
		if ctx.MouseClicked {
			ctx.ActiveID = id
		}
	}

	// Determine visual state
	var boxColor graphics.Color
	if ctx.HotID == id {
		boxColor = ctx.Style.ButtonHoverColor
	} else {
		boxColor = ctx.Style.FrameBgColor
	}

	// Draw checkbox frame with depth
	graphics.FillRect(w, graphics.NewRect(x, y, boxSize, boxSize), boxColor)
	// Dark inner edge
	graphics.DrawLine(w, x, y, x+boxSize-1, y, graphics.NewColor(0, 0, 0, 100))
	graphics.DrawLine(w, x, y, x, y+boxSize-1, graphics.NewColor(0, 0, 0, 100))

	// Draw checkmark if checked (ImGui-style tick)
	if value {
		checkColor := ctx.Style.CheckmarkColor
		// Draw a checkmark (tick shape)
		cx := x + boxSize/2
		cy := y + boxSize/2
		// Left part of tick (going down-right)
		graphics.DrawLine(w, cx-5, cy, cx-2, cy+4, checkColor)
		graphics.DrawLine(w, cx-4, cy, cx-1, cy+4, checkColor)
		// Right part of tick (going up-right)
		graphics.DrawLine(w, cx-2, cy+4, cx+5, cy-3, checkColor)
		graphics.DrawLine(w, cx-1, cy+4, cx+6, cy-3, checkColor)
	}

	// Draw label
	labelX := x + boxSize + ctx.Style.Padding
	labelY := y + (boxSize-TextHeight(ctx.Style.FontSize))/2
	DrawText(w, label, labelX, labelY, ctx.Style.FontSize, ctx.Style.TextColor)

	// Toggle on click
	newValue := value
	if ctx.ReleasedID == id && ctx.MouseReleased && hovered {
		newValue = !value
	}
	return ctx, newValue
}

// Slider draws a horizontal slider and returns updated context and new value
// value and result are in range [min, max]
func Slider(ctx GuiContext, w graphics.Window, label string, x int32, y int32, width int32, min float64, max float64, value float64) (GuiContext, float64) {
	id := GenID(label)
	height := ctx.Style.SliderHeight
	grabW := int32(12) // Smaller grab handle like ImGui

	// Draw label to the left
	labelW := TextWidth(label, ctx.Style.FontSize)
	labelY := y + (height-TextHeight(ctx.Style.FontSize))/2
	DrawText(w, label, x, labelY, ctx.Style.FontSize, ctx.Style.TextColor)

	// Slider track starts after label
	trackX := x + labelW + ctx.Style.Padding
	trackW := width - labelW - ctx.Style.Padding

	// Clamp value
	if value < min {
		value = min
	}
	if value > max {
		value = max
	}

	// Calculate grab position
	rangeVal := max - min
	if rangeVal == 0 {
		rangeVal = 1
	}
	t := (value - min) / rangeVal
	grabRange := trackW - grabW
	grabX := trackX + int32(float64(grabRange)*t)

	// Hit test (entire slider track area)
	hovered := pointInRect(ctx.MouseX, ctx.MouseY, trackX, y, trackW, height)

	// Update hot/active state
	if hovered {
		ctx.HotID = id
		if ctx.MouseClicked {
			ctx.ActiveID = id
		}
	}

	// Draw track background (full width, ImGui style)
	graphics.FillRect(w, graphics.NewRect(trackX, y, trackW, height), ctx.Style.FrameBgColor)
	// Dark inner edge for depth
	graphics.DrawLine(w, trackX, y, trackX+trackW-1, y, graphics.NewColor(0, 0, 0, 100))
	graphics.DrawLine(w, trackX, y, trackX, y+height-1, graphics.NewColor(0, 0, 0, 100))

	// Draw filled portion (from start to grab)
	fillW := grabX - trackX + grabW/2
	if fillW > 0 {
		graphics.FillRect(w, graphics.NewRect(trackX+1, y+1, fillW, height-2), ctx.Style.SliderTrackColor)
	}

	// Draw grab handle
	var grabColor graphics.Color
	if ctx.ActiveID == id {
		grabColor = ctx.Style.ButtonActiveColor
	} else if ctx.HotID == id {
		grabColor = ctx.Style.ButtonHoverColor
	} else {
		grabColor = ctx.Style.SliderKnobColor
	}
	graphics.FillRect(w, graphics.NewRect(grabX, y, grabW, height), grabColor)

	// Handle dragging
	if ctx.ActiveID == id && ctx.MouseDown {
		// Calculate new value from mouse position
		mouseT := float64(ctx.MouseX-trackX-grabW/2) / float64(grabRange)
		if mouseT < 0 {
			mouseT = 0
		}
		if mouseT > 1 {
			mouseT = 1
		}
		value = min + mouseT*rangeVal
	}

	return ctx, value
}

// --- Panels and Frames ---

// Panel draws a panel/window background with title
func Panel(ctx GuiContext, w graphics.Window, title string, x int32, y int32, width int32, height int32) {
	// Draw title bar with blue ImGui style
	titleH := TextHeight(ctx.Style.FontSize) + ctx.Style.Padding*2
	graphics.FillRect(w, graphics.NewRect(x, y, width, titleH), ctx.Style.TitleBgColor)

	// Title text centered vertically, left padded
	DrawText(w, title, x+ctx.Style.Padding, y+(titleH-TextHeight(ctx.Style.FontSize))/2, ctx.Style.FontSize, ctx.Style.TextColor)

	// Draw panel body - dark gray/black
	graphics.FillRect(w, graphics.NewRect(x, y+titleH, width, height-titleH), ctx.Style.BackgroundColor)

	// Outer border for the whole window
	graphics.DrawRect(w, graphics.NewRect(x, y, width, height), ctx.Style.BorderColor)

	// Subtle inner highlight on title bar bottom
	graphics.DrawLine(w, x+1, y+titleH-1, x+width-2, y+titleH-1, graphics.NewColor(255, 255, 255, 20))
}

// DraggablePanel draws a draggable panel/window and returns updated context and window state
func DraggablePanel(ctx GuiContext, w graphics.Window, title string, state WindowState) (GuiContext, WindowState) {
	// Generate ID from title (avoid reusing title after concat)
	idStr := title
	idStr += "_panel"
	id := GenID(idStr)
	titleH := TextHeight(ctx.Style.FontSize) + ctx.Style.Padding*2

	// Check if mouse is in title bar (drag area)
	inTitleBar := pointInRect(ctx.MouseX, ctx.MouseY, state.X, state.Y, state.Width, titleH)

	// Handle drag start
	if inTitleBar && ctx.MouseClicked {
		state.Dragging = true
		state.DragOffsetX = ctx.MouseX - state.X
		state.DragOffsetY = ctx.MouseY - state.Y
		ctx.ActiveID = id
	}

	// Handle dragging
	if state.Dragging && ctx.MouseDown {
		state.X = ctx.MouseX - state.DragOffsetX
		state.Y = ctx.MouseY - state.DragOffsetY
	}

	// Handle drag end
	if state.Dragging && ctx.MouseReleased {
		state.Dragging = false
		if ctx.ActiveID == id {
			ctx.ActiveID = 0
		}
	}

	// Draw the panel at current position
	Panel(ctx, w, title, state.X, state.Y, state.Width, state.Height)

	return ctx, state
}

// Separator draws a horizontal separator line
func Separator(ctx GuiContext, w graphics.Window, x int32, y int32, width int32) {
	graphics.DrawLine(w, x, y, x+width, y, ctx.Style.BorderColor)
}

// --- Menu System ---

// BeginMenuBar starts a menu bar at the given position
func BeginMenuBar(ctx GuiContext, w graphics.Window, state MenuState, x int32, y int32, width int32) (GuiContext, MenuState) {
	height := TextHeight(ctx.Style.FontSize) + ctx.Style.Padding*2

	// Draw menu bar background
	graphics.FillRect(w, graphics.NewRect(x, y, width, height), ctx.Style.TitleBgColor)
	// Bottom border
	graphics.DrawLine(w, x, y+height-1, x+width, y+height-1, ctx.Style.BorderColor)

	// Store menu bar info for dropdown positioning
	state.MenuBarX = x
	state.MenuBarY = y
	state.MenuBarH = height
	state.CurrentMenuX = x + ctx.Style.Padding

	// Check for click outside any menu to close
	state.ClickedOutside = ctx.MouseClicked

	return ctx, state
}

// EndMenuBar finishes the menu bar and handles click-outside-to-close
func EndMenuBar(ctx GuiContext, state MenuState) (GuiContext, MenuState) {
	// If clicked outside and no menu item was hit, close menu
	if state.ClickedOutside && state.OpenMenuID != 0 {
		state.OpenMenuID = 0
	}
	return ctx, state
}

// Menu draws a menu header and returns true if the dropdown should be shown
func Menu(ctx GuiContext, w graphics.Window, state MenuState, label string) (GuiContext, MenuState, bool) {
	id := GenID(label)
	padding := ctx.Style.Padding
	textW := TextWidth(label, ctx.Style.FontSize)
	textH := TextHeight(ctx.Style.FontSize)
	menuW := textW + padding*2
	menuH := state.MenuBarH

	x := state.CurrentMenuX
	y := state.MenuBarY

	// Hit test
	hovered := pointInRect(ctx.MouseX, ctx.MouseY, x, y, menuW, menuH)

	// Determine if this menu is open
	isOpen := state.OpenMenuID == id

	// Draw background if hovered or open
	if hovered || isOpen {
		graphics.FillRect(w, graphics.NewRect(x, y, menuW, menuH-1), ctx.Style.ButtonHoverColor)
		// If clicked, toggle menu
		if ctx.MouseClicked {
			if isOpen {
				state.OpenMenuID = 0
				isOpen = false
			} else {
				state.OpenMenuID = id
				isOpen = true
			}
			state.ClickedOutside = false // Click was on menu, not outside
		}
	}

	// Draw label centered vertically
	textY := y + (menuH-textH)/2
	DrawText(w, label, x+padding, textY, ctx.Style.FontSize, ctx.Style.TextColor)

	// Store position for dropdown and advance for next menu
	state.CurrentMenuW = menuW
	state.CurrentMenuX = x + menuW

	return ctx, state, isOpen
}

// BeginDropdown starts a dropdown menu area, returns the dropdown Y position
func BeginDropdown(ctx GuiContext, w graphics.Window, state MenuState) (GuiContext, int32) {
	// Dropdown appears below menu bar, aligned with current menu header
	dropY := state.MenuBarY + state.MenuBarH
	return ctx, dropY
}

// MenuItem draws a menu item and returns true if clicked
func MenuItem(ctx GuiContext, w graphics.Window, state MenuState, label string, dropX int32, dropY int32, itemIndex int32) (GuiContext, MenuState, bool) {
	padding := ctx.Style.Padding
	textH := TextHeight(ctx.Style.FontSize)
	itemH := textH + padding*2
	itemW := int32(150) // Fixed width for menu items

	y := dropY + itemIndex*itemH

	// Draw item background
	graphics.FillRect(w, graphics.NewRect(dropX, y, itemW, itemH), ctx.Style.BackgroundColor)

	// Hit test
	hovered := pointInRect(ctx.MouseX, ctx.MouseY, dropX, y, itemW, itemH)
	clicked := false

	if hovered {
		graphics.FillRect(w, graphics.NewRect(dropX, y, itemW, itemH), ctx.Style.ButtonHoverColor)
		state.ClickedOutside = false // Click is on menu item
		if ctx.MouseClicked {
			clicked = true
			state.OpenMenuID = 0 // Close menu after click
		}
	}

	// Draw label
	textY := y + (itemH-textH)/2
	DrawText(w, label, dropX+padding, textY, ctx.Style.FontSize, ctx.Style.TextColor)

	// Draw border
	graphics.DrawRect(w, graphics.NewRect(dropX, dropY, itemW, (itemIndex+1)*itemH), ctx.Style.BorderColor)

	return ctx, state, clicked
}

// MenuItemSeparator draws a separator line in a dropdown menu
func MenuItemSeparator(ctx GuiContext, w graphics.Window, dropX int32, dropY int32, itemIndex int32) {
	padding := ctx.Style.Padding
	textH := TextHeight(ctx.Style.FontSize)
	itemH := textH + padding*2
	itemW := int32(150)

	y := dropY + itemIndex*itemH + itemH/2

	graphics.FillRect(w, graphics.NewRect(dropX, dropY+itemIndex*itemH, itemW, itemH), ctx.Style.BackgroundColor)
	graphics.DrawLine(w, dropX+padding, y, dropX+itemW-padding, y, ctx.Style.BorderColor)
}

// --- Layout Helpers ---

// BeginLayout starts auto-layout at the given position
func BeginLayout(ctx GuiContext, x int32, y int32, spacing int32) GuiContext {
	ctx.CursorX = x
	ctx.CursorY = y
	ctx.Spacing = spacing
	return ctx
}

// NextRow moves cursor to next row
func NextRow(ctx GuiContext, height int32) GuiContext {
	ctx.CursorY = ctx.CursorY + height + ctx.Spacing
	return ctx
}

// AutoLabel draws label at cursor position and advances cursor
func AutoLabel(ctx GuiContext, w graphics.Window, text string) GuiContext {
	Label(ctx, w, text, ctx.CursorX, ctx.CursorY)
	ctx = NextRow(ctx, TextHeight(ctx.Style.FontSize))
	return ctx
}

// AutoButton draws button at cursor position and advances cursor
func AutoButton(ctx GuiContext, w graphics.Window, label string, width int32, height int32) (GuiContext, bool) {
	var result bool
	ctx, result = Button(ctx, w, label, ctx.CursorX, ctx.CursorY, width, height)
	ctx = NextRow(ctx, height)
	return ctx, result
}

// AutoCheckbox draws checkbox at cursor position and advances cursor
func AutoCheckbox(ctx GuiContext, w graphics.Window, label string, value bool) (GuiContext, bool) {
	var result bool
	ctx, result = Checkbox(ctx, w, label, ctx.CursorX, ctx.CursorY, value)
	ctx = NextRow(ctx, ctx.Style.CheckboxSize)
	return ctx, result
}

// AutoSlider draws slider at cursor position and advances cursor
func AutoSlider(ctx GuiContext, w graphics.Window, label string, width int32, min float64, max float64, value float64) (GuiContext, float64) {
	var result float64
	ctx, result = Slider(ctx, w, label, ctx.CursorX, ctx.CursorY, width, min, max, value)
	ctx = NextRow(ctx, ctx.Style.SliderHeight)
	return ctx, result
}

// --- Embedded 8x8 Bitmap Font ---
// Each character is 8 bytes, where each byte represents one row
// Bit 7 is leftmost pixel, bit 0 is rightmost pixel

func getFontData() []uint8 {
	return []uint8{
	// Character 32: ' ' (space)
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// Character 33: '!'
	0x18, 0x18, 0x18, 0x18, 0x18, 0x00, 0x18, 0x00,
	// Character 34: '"'
	0x6C, 0x6C, 0x24, 0x00, 0x00, 0x00, 0x00, 0x00,
	// Character 35: '#'
	0x6C, 0x6C, 0xFE, 0x6C, 0xFE, 0x6C, 0x6C, 0x00,
	// Character 36: '$'
	0x18, 0x3E, 0x60, 0x3C, 0x06, 0x7C, 0x18, 0x00,
	// Character 37: '%'
	0x00, 0xC6, 0xCC, 0x18, 0x30, 0x66, 0xC6, 0x00,
	// Character 38: '&'
	0x38, 0x6C, 0x38, 0x76, 0xDC, 0xCC, 0x76, 0x00,
	// Character 39: '''
	0x18, 0x18, 0x30, 0x00, 0x00, 0x00, 0x00, 0x00,
	// Character 40: '('
	0x0C, 0x18, 0x30, 0x30, 0x30, 0x18, 0x0C, 0x00,
	// Character 41: ')'
	0x30, 0x18, 0x0C, 0x0C, 0x0C, 0x18, 0x30, 0x00,
	// Character 42: '*'
	0x00, 0x66, 0x3C, 0xFF, 0x3C, 0x66, 0x00, 0x00,
	// Character 43: '+'
	0x00, 0x18, 0x18, 0x7E, 0x18, 0x18, 0x00, 0x00,
	// Character 44: ','
	0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x30,
	// Character 45: '-'
	0x00, 0x00, 0x00, 0x7E, 0x00, 0x00, 0x00, 0x00,
	// Character 46: '.'
	0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x00,
	// Character 47: '/'
	0x06, 0x0C, 0x18, 0x30, 0x60, 0xC0, 0x80, 0x00,
	// Character 48: '0'
	0x7C, 0xC6, 0xCE, 0xD6, 0xE6, 0xC6, 0x7C, 0x00,
	// Character 49: '1'
	0x18, 0x38, 0x18, 0x18, 0x18, 0x18, 0x7E, 0x00,
	// Character 50: '2'
	0x7C, 0xC6, 0x06, 0x1C, 0x30, 0x66, 0xFE, 0x00,
	// Character 51: '3'
	0x7C, 0xC6, 0x06, 0x3C, 0x06, 0xC6, 0x7C, 0x00,
	// Character 52: '4'
	0x1C, 0x3C, 0x6C, 0xCC, 0xFE, 0x0C, 0x1E, 0x00,
	// Character 53: '5'
	0xFE, 0xC0, 0xC0, 0xFC, 0x06, 0xC6, 0x7C, 0x00,
	// Character 54: '6'
	0x38, 0x60, 0xC0, 0xFC, 0xC6, 0xC6, 0x7C, 0x00,
	// Character 55: '7'
	0xFE, 0xC6, 0x0C, 0x18, 0x30, 0x30, 0x30, 0x00,
	// Character 56: '8'
	0x7C, 0xC6, 0xC6, 0x7C, 0xC6, 0xC6, 0x7C, 0x00,
	// Character 57: '9'
	0x7C, 0xC6, 0xC6, 0x7E, 0x06, 0x0C, 0x78, 0x00,
	// Character 58: ':'
	0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x00,
	// Character 59: ';'
	0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x30,
	// Character 60: '<'
	0x06, 0x0C, 0x18, 0x30, 0x18, 0x0C, 0x06, 0x00,
	// Character 61: '='
	0x00, 0x00, 0x7E, 0x00, 0x00, 0x7E, 0x00, 0x00,
	// Character 62: '>'
	0x60, 0x30, 0x18, 0x0C, 0x18, 0x30, 0x60, 0x00,
	// Character 63: '?'
	0x7C, 0xC6, 0x0C, 0x18, 0x18, 0x00, 0x18, 0x00,
	// Character 64: '@'
	0x7C, 0xC6, 0xDE, 0xDE, 0xDE, 0xC0, 0x78, 0x00,
	// Character 65: 'A'
	0x38, 0x6C, 0xC6, 0xFE, 0xC6, 0xC6, 0xC6, 0x00,
	// Character 66: 'B'
	0xFC, 0x66, 0x66, 0x7C, 0x66, 0x66, 0xFC, 0x00,
	// Character 67: 'C'
	0x3C, 0x66, 0xC0, 0xC0, 0xC0, 0x66, 0x3C, 0x00,
	// Character 68: 'D'
	0xF8, 0x6C, 0x66, 0x66, 0x66, 0x6C, 0xF8, 0x00,
	// Character 69: 'E'
	0xFE, 0x62, 0x68, 0x78, 0x68, 0x62, 0xFE, 0x00,
	// Character 70: 'F'
	0xFE, 0x62, 0x68, 0x78, 0x68, 0x60, 0xF0, 0x00,
	// Character 71: 'G'
	0x3C, 0x66, 0xC0, 0xC0, 0xCE, 0x66, 0x3A, 0x00,
	// Character 72: 'H'
	0xC6, 0xC6, 0xC6, 0xFE, 0xC6, 0xC6, 0xC6, 0x00,
	// Character 73: 'I'
	0x3C, 0x18, 0x18, 0x18, 0x18, 0x18, 0x3C, 0x00,
	// Character 74: 'J'
	0x1E, 0x0C, 0x0C, 0x0C, 0xCC, 0xCC, 0x78, 0x00,
	// Character 75: 'K'
	0xE6, 0x66, 0x6C, 0x78, 0x6C, 0x66, 0xE6, 0x00,
	// Character 76: 'L'
	0xF0, 0x60, 0x60, 0x60, 0x62, 0x66, 0xFE, 0x00,
	// Character 77: 'M'
	0xC6, 0xEE, 0xFE, 0xFE, 0xD6, 0xC6, 0xC6, 0x00,
	// Character 78: 'N'
	0xC6, 0xE6, 0xF6, 0xDE, 0xCE, 0xC6, 0xC6, 0x00,
	// Character 79: 'O'
	0x7C, 0xC6, 0xC6, 0xC6, 0xC6, 0xC6, 0x7C, 0x00,
	// Character 80: 'P'
	0xFC, 0x66, 0x66, 0x7C, 0x60, 0x60, 0xF0, 0x00,
	// Character 81: 'Q'
	0x7C, 0xC6, 0xC6, 0xC6, 0xD6, 0xDE, 0x7C, 0x06,
	// Character 82: 'R'
	0xFC, 0x66, 0x66, 0x7C, 0x6C, 0x66, 0xE6, 0x00,
	// Character 83: 'S'
	0x7C, 0xC6, 0x60, 0x38, 0x0C, 0xC6, 0x7C, 0x00,
	// Character 84: 'T'
	0x7E, 0x7E, 0x5A, 0x18, 0x18, 0x18, 0x3C, 0x00,
	// Character 85: 'U'
	0xC6, 0xC6, 0xC6, 0xC6, 0xC6, 0xC6, 0x7C, 0x00,
	// Character 86: 'V'
	0xC6, 0xC6, 0xC6, 0xC6, 0xC6, 0x6C, 0x38, 0x00,
	// Character 87: 'W'
	0xC6, 0xC6, 0xC6, 0xD6, 0xD6, 0xFE, 0x6C, 0x00,
	// Character 88: 'X'
	0xC6, 0xC6, 0x6C, 0x38, 0x6C, 0xC6, 0xC6, 0x00,
	// Character 89: 'Y'
	0x66, 0x66, 0x66, 0x3C, 0x18, 0x18, 0x3C, 0x00,
	// Character 90: 'Z'
	0xFE, 0xC6, 0x8C, 0x18, 0x32, 0x66, 0xFE, 0x00,
	// Character 91: '['
	0x3C, 0x30, 0x30, 0x30, 0x30, 0x30, 0x3C, 0x00,
	// Character 92: '\'
	0xC0, 0x60, 0x30, 0x18, 0x0C, 0x06, 0x02, 0x00,
	// Character 93: ']'
	0x3C, 0x0C, 0x0C, 0x0C, 0x0C, 0x0C, 0x3C, 0x00,
	// Character 94: '^'
	0x10, 0x38, 0x6C, 0xC6, 0x00, 0x00, 0x00, 0x00,
	// Character 95: '_'
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF,
	// Character 96: '`'
	0x30, 0x18, 0x0C, 0x00, 0x00, 0x00, 0x00, 0x00,
	// Character 97: 'a'
	0x00, 0x00, 0x78, 0x0C, 0x7C, 0xCC, 0x76, 0x00,
	// Character 98: 'b'
	0xE0, 0x60, 0x7C, 0x66, 0x66, 0x66, 0xDC, 0x00,
	// Character 99: 'c'
	0x00, 0x00, 0x7C, 0xC6, 0xC0, 0xC6, 0x7C, 0x00,
	// Character 100: 'd'
	0x1C, 0x0C, 0x7C, 0xCC, 0xCC, 0xCC, 0x76, 0x00,
	// Character 101: 'e'
	0x00, 0x00, 0x7C, 0xC6, 0xFE, 0xC0, 0x7C, 0x00,
	// Character 102: 'f'
	0x3C, 0x66, 0x60, 0xF8, 0x60, 0x60, 0xF0, 0x00,
	// Character 103: 'g'
	0x00, 0x00, 0x76, 0xCC, 0xCC, 0x7C, 0x0C, 0xF8,
	// Character 104: 'h'
	0xE0, 0x60, 0x6C, 0x76, 0x66, 0x66, 0xE6, 0x00,
	// Character 105: 'i'
	0x18, 0x00, 0x38, 0x18, 0x18, 0x18, 0x3C, 0x00,
	// Character 106: 'j'
	0x06, 0x00, 0x06, 0x06, 0x06, 0x66, 0x66, 0x3C,
	// Character 107: 'k'
	0xE0, 0x60, 0x66, 0x6C, 0x78, 0x6C, 0xE6, 0x00,
	// Character 108: 'l'
	0x38, 0x18, 0x18, 0x18, 0x18, 0x18, 0x3C, 0x00,
	// Character 109: 'm'
	0x00, 0x00, 0xEC, 0xFE, 0xD6, 0xD6, 0xD6, 0x00,
	// Character 110: 'n'
	0x00, 0x00, 0xDC, 0x66, 0x66, 0x66, 0x66, 0x00,
	// Character 111: 'o'
	0x00, 0x00, 0x7C, 0xC6, 0xC6, 0xC6, 0x7C, 0x00,
	// Character 112: 'p'
	0x00, 0x00, 0xDC, 0x66, 0x66, 0x7C, 0x60, 0xF0,
	// Character 113: 'q'
	0x00, 0x00, 0x76, 0xCC, 0xCC, 0x7C, 0x0C, 0x1E,
	// Character 114: 'r'
	0x00, 0x00, 0xDC, 0x76, 0x60, 0x60, 0xF0, 0x00,
	// Character 115: 's'
	0x00, 0x00, 0x7E, 0xC0, 0x7C, 0x06, 0xFC, 0x00,
	// Character 116: 't'
	0x30, 0x30, 0xFC, 0x30, 0x30, 0x36, 0x1C, 0x00,
	// Character 117: 'u'
	0x00, 0x00, 0xCC, 0xCC, 0xCC, 0xCC, 0x76, 0x00,
	// Character 118: 'v'
	0x00, 0x00, 0xC6, 0xC6, 0xC6, 0x6C, 0x38, 0x00,
	// Character 119: 'w'
	0x00, 0x00, 0xC6, 0xD6, 0xD6, 0xFE, 0x6C, 0x00,
	// Character 120: 'x'
	0x00, 0x00, 0xC6, 0x6C, 0x38, 0x6C, 0xC6, 0x00,
	// Character 121: 'y'
	0x00, 0x00, 0xC6, 0xC6, 0xC6, 0x7E, 0x06, 0xFC,
	// Character 122: 'z'
	0x00, 0x00, 0xFE, 0x8C, 0x18, 0x32, 0xFE, 0x00,
	// Character 123: '{'
	0x0E, 0x18, 0x18, 0x70, 0x18, 0x18, 0x0E, 0x00,
	// Character 124: '|'
	0x18, 0x18, 0x18, 0x18, 0x18, 0x18, 0x18, 0x00,
	// Character 125: '}'
	0x70, 0x18, 0x18, 0x0E, 0x18, 0x18, 0x70, 0x00,
	// Character 126: '~'
	0x76, 0xDC, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	// Character 127: DEL (block character)
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	}
}
