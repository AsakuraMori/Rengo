package engine

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type ScriptEngine struct {
	engine           *Engine
	waitingForChoice bool
	waitingForInput  bool
	variables        map[string]interface{}
	currentText      string
	textReady        bool
	textQueue        []string
	choicesToShow    []Choice
	pc               int
	scriptLines      []string
	currentLine      int
	affection        map[string]int // 好感度系统
	pendingJump      string         // 用于存储待执行的跳转目标
	conditionalStack []bool         // 用于跟踪条件分支的状态
}

func NewScriptEngine(engine *Engine) *ScriptEngine {
	se := &ScriptEngine{
		engine:           engine,
		waitingForChoice: false,
		variables:        make(map[string]interface{}),
		textQueue:        make([]string, 0),
		waitingForInput:  false,
		scriptLines:      make([]string, 0),
		currentLine:      0,
		affection:        make(map[string]int), // 初始化好感度系统
	}
	return se
}

// 加载KAG脚本并逐行解析
func (se *ScriptEngine) LoadScript(scriptName string) error {
	file, err := os.Open(scriptName)
	if err != nil {
		return fmt.Errorf("failed to read script file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "--") {
			continue // 跳过空行和注释
		}
		se.scriptLines = append(se.scriptLines, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read script file: %v", err)
	}

	se.currentLine = 0
	return nil
}

// 逐行执行脚本
func (se *ScriptEngine) ExecuteStep() bool {
	if se.engine.ChoiceSystem.IsWaitingForChoice() {
		selected, jumpTo := se.engine.ChoiceSystem.HandleInput()
		if selected {
			se.clearChoices() // 清除选项
			se.jumpToLabel(jumpTo)
			// 立即执行跳转后的所有命令
			for se.currentLine < len(se.scriptLines) {
				if !se.ExecuteStep() {
					break
				}
			}
			return true
		}
	}
	if se.waitingForInput {
		// 等待用户输入
		return false
	}
	if se.pendingJump != "" {
		se.jumpToLabel(se.pendingJump)
		se.pendingJump = "" // 清除待跳转
		return true
	}
	// 检查是否已经执行完所有行
	if se.currentLine >= len(se.scriptLines) {
		return false
	}

	// 获取当前行
	line := se.scriptLines[se.currentLine]
	se.currentLine++

	// 解析并执行当前行
	if strings.HasPrefix(line, "@") {
		se.parseCommand(line)
	} else if strings.HasPrefix(line, ":") {
		// 标签，暂时忽略
	} else {
		// 文本行
		se.currentText = line
		se.engine.TextDisplay.SetText(line) // 设置文字内容
		se.waitingForInput = true
		return true
	}
	if len(se.choicesToShow) > 0 {
		se.showChoices()
		return true
	}
	return true
}

// 显示选项并等待用户选择
func (se *ScriptEngine) showChoices() {
	if len(se.choicesToShow) == 0 {
		se.waitingForChoice = false
		se.engine.ChoiceSystem.IsActive = false
		return
	}
	se.engine.ChoiceSystem.SetChoices(se.choicesToShow, se.engine.Width)
	se.waitingForChoice = true
	se.engine.ChoiceSystem.IsActive = true
}

func (se *ScriptEngine) clearChoices() {
	se.choicesToShow = []Choice{}
	se.engine.ChoiceSystem.SetChoices(se.choicesToShow, se.engine.Width)
	se.waitingForChoice = false
	se.engine.ChoiceSystem.IsActive = false
}

// 跳转到指定标签
func (se *ScriptEngine) jumpToLabel(label string) {
	se.clearChoices() // 清除选项
	for i, line := range se.scriptLines {
		if strings.HasPrefix(line, ":") && strings.TrimSpace(line[1:]) == label {
			se.currentLine = i + 1 // 跳转到标签的下一行
			return
		}
	}
	log.Printf("未找到标签: %s", label)
}

// 解析命令
func (se *ScriptEngine) parseCommand(line string) {
	parts := strings.Fields(line[1:])
	command := parts[0]
	args := parts[1:]

	switch command {
	case "bg":
		se.handleBackgroundCommand(args)
	case "chara":
		se.handleCharacterCommand(args)
	case "choice":
		se.handleChoiceCommand(args)
	case "affection":
		se.handleAffectionCommand(args)
	case "if", "elseif", "else":
		se.handleIfCommand(parts)
	case "endif":
		if len(se.conditionalStack) > 0 {
			se.conditionalStack = se.conditionalStack[:len(se.conditionalStack)-1]
		}
	case "jump":
		se.handleJumpCommand(args)
	case "clear":
		se.handleClearLayer(args)
	default:
		log.Printf("未知命令: %s", command)
	}
}

// 处理背景命令

func (se *ScriptEngine) handleClearLayer(args []string) {
	idx, _ := strconv.Atoi(args[0])
	err := se.engine.ClearLayer(idx)
	if err != nil {
		log.Printf("清除背景失败: %v", err)
	}
}

func (se *ScriptEngine) effectCommand(command string, args []string) {
	switch command {
	case "transition":
		layerIndex, _ := strconv.Atoi(args[0])
		targetImage := args[1]
		maskImage := args[2]
		//speed, _ := strconv.ParseFloat(args[3], 64)

		// 获取当前图层图像作为源
		source := se.engine.Layers[layerIndex].ImageDisplay.current

		// 使用ImageDisplay的LoadImage方法加载目标图像
		targetDisplay := se.engine.Layers[layerIndex].ImageDisplay
		if err := targetDisplay.LoadImage(targetImage, targetImage); err != nil {
			log.Printf("加载目标图像失败: %v", err)
			return
		}
		target := targetDisplay.images[targetImage]

		// 添加渐变效果
		err := se.engine.EffectSystem.AddTransitionEffect(
			layerIndex,
			source,
			target,
			maskImage,
			0.01,
		)
		if err != nil {
			log.Printf("添加渐变效果失败: %v", err)
		}

		// 设置目标图像为图层的新图像
		//se.engine.Layers[layerIndex].ImageDisplay.current = target
	}
}
func (se *ScriptEngine) handleBackgroundCommand(args []string) {
	idx, _ := strconv.Atoi(args[0])
	imagePath := args[1]
	if len(args) > 2 {
		fadeCmd := args[3]
		se.effectCommand(fadeCmd, args)
	}
	err := se.engine.SetLayerImage(idx, "background", imagePath)
	if err != nil {
		log.Printf("设置背景失败: %v", err)
	} else {
		log.Printf("设置背景: %s", imagePath)
	}
}

// 处理立绘命令
func (se *ScriptEngine) handleCharacterCommand(args []string) {
	idx, _ := strconv.Atoi(args[0])
	position := args[1]
	imagePath := args[2]
	err := se.engine.SetLayerCharacter(idx, position, imagePath)
	if err != nil {
		log.Printf("设置立绘失败: %v", err)
	} else {
		log.Printf("设置立绘: %s -> %s", position, imagePath)
	}
}

// 处理选择支命令
func (se *ScriptEngine) handleChoiceCommand(args []string) {
	se.clearChoices() // 清除旧的选项
	for i := 0; i < len(args); i += 3 {
		if i+2 >= len(args) {
			log.Printf("选项格式错误: %v", args)
			break
		}
		text := args[i]
		arrow := args[i+1]
		jumpTo := args[i+2]

		if arrow != "->" {
			log.Printf("选项格式错误: 缺少 ->")
			break
		}

		se.choicesToShow = append(se.choicesToShow, Choice{
			Text:   text,
			JumpTo: jumpTo,
		})
	}
	se.engine.ChoiceSystem.SetChoices(se.choicesToShow, se.engine.Width)
	se.waitingForChoice = true
	se.engine.ChoiceSystem.IsActive = true
	log.Printf("设置选择支: %v", se.choicesToShow)
}

// 处理好感度命令
func (se *ScriptEngine) handleAffectionCommand(args []string) {
	character := args[0]
	value := args[1]
	delta, _ := strconv.Atoi(value)
	se.affection[character] += delta
	log.Printf("好感度变化: %s +%d", character, delta)
}

func (se *ScriptEngine) handleIfCommand(args []string) {
	if args[0] == "else" {
		// 处理 @else
		if len(se.conditionalStack) > 0 && !se.conditionalStack[len(se.conditionalStack)-1] {
			se.conditionalStack[len(se.conditionalStack)-1] = true
		} else {
			se.skipToEndif()
		}
		return
	}

	var conditionMet bool
	if args[0] == "elseif" {
		// 处理 @elseif
		if len(se.conditionalStack) > 0 && !se.conditionalStack[len(se.conditionalStack)-1] {
			args = args[1:] // 移除 "elseif" 参数
			conditionMet = se.evaluateCondition(args)
			se.conditionalStack[len(se.conditionalStack)-1] = conditionMet
		} else {
			se.skipToEndif()
			return
		}
	} else {
		// 处理 @if
		conditionMet = se.evaluateCondition(args)
		se.conditionalStack = append(se.conditionalStack, conditionMet)
	}

	if !conditionMet {
		se.skipToNextBranch()
	}
}

func (se *ScriptEngine) evaluateCondition(args []string) bool {
	// 解析条件：affection Yuki >= 5
	character := args[2]
	operator := args[3]
	targetValue, _ := strconv.Atoi(args[4])

	currentValue := se.affection[character]

	switch operator {
	case ">=":
		return currentValue >= targetValue
	case "<=":
		return currentValue <= targetValue
	case ">":
		return currentValue > targetValue
	case "<":
		return currentValue < targetValue
	case "==":
		return currentValue == targetValue
	}
	return false
}

func (se *ScriptEngine) skipToNextBranch() {
	for se.currentLine < len(se.scriptLines) {
		line := se.scriptLines[se.currentLine]
		if strings.HasPrefix(line, "@elseif") || strings.HasPrefix(line, "@else") || strings.HasPrefix(line, "@endif") {
			return
		}
		se.currentLine++
	}
}

func (se *ScriptEngine) skipToEndif() {
	for se.currentLine < len(se.scriptLines) {
		line := se.scriptLines[se.currentLine]
		if strings.HasPrefix(line, "@endif") {
			return
		}
		se.currentLine++
	}
}

// 处理跳转命令
func (se *ScriptEngine) handleJumpCommand(args []string) {
	jumpTo := args[0]
	se.pendingJump = jumpTo
	log.Printf("设置跳转到: %s", jumpTo)
}
