package main

import (
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	img            *ebiten.Image
	count          int
	winX, winY     int
	dragging       bool
	startX, startY int
	showMenu       bool
	menuX, menuY   int

	// 互动相关状态
	isHovering  bool
	jumpOffset  float64 // 点击跳跃的偏移量
	targetAngle float64 // 盯着鼠标看的角度
}

func (g *Game) Update() error {
	g.count++
	cx, cy := ebiten.CursorPosition()

	// 1. 基础范围判定 (是否鼠标在宠物身上)
	g.isHovering = cx >= 10 && cx <= 150 && cy >= 10 && cy <= 150

	// 2. 菜单逻辑
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		g.showMenu, g.menuX, g.menuY = true, cx, cy
	}
	if g.showMenu {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			if cx >= g.menuX && cx <= g.menuX+60 && cy >= g.menuY && cy <= g.menuY+30 {
				os.Exit(0)
			}
			g.showMenu = false
		}
		return nil
	}

	// 3. 眼神/身体跟随逻辑 (计算指向鼠标的角度)
	if !g.dragging {
		// 计算光标相对于中心 (80,80) 的角度
		angle := math.Atan2(float64(cy-80), float64(cx-80))
		g.targetAngle = angle * 0.1 // 轻轻倾斜，不要转得太夸张
	}

	// 4. 左键逻辑 (拖拽 + 点击跳跃)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if !g.dragging && g.isHovering {
			g.dragging, g.startX, g.startY = true, cx, cy
			g.jumpOffset = -15 // 触发跳跃感
		} else if g.dragging {
			g.winX += cx - g.startX
			g.winY += cy - g.startY
			ebiten.SetWindowPosition(g.winX, g.winY)
		}
	} else {
		g.dragging = false
	}

	// 5. 跳跃动画衰减
	g.jumpOffset *= 0.9 // 逐渐恢复原位
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.img == nil { return }

	op := &ebiten.DrawImageOptions{}
	
	// 动态调整动画速度：悬浮时呼吸变快
	animSpeed := 0.05
	if g.isHovering { animSpeed = 0.15 }
	
	t := float64(g.count) * animSpeed
	
	// 基础变换
	s := 1.0 + math.Sin(t)*0.02
	if g.isHovering { s += 0.05 } // 悬浮时微微放大
	
	w, h := g.img.Bounds().Dx(), g.img.Bounds().Dy()
	
	// 变换序列
	op.GeoM.Translate(-float64(w)/2, -float64(h)/2) // 移到中心
	op.GeoM.Rotate(g.targetAngle)                   // 盯着鼠标看
	op.GeoM.Scale(140.0/float64(w)*s, 140.0/float64(h)*s)
	op.GeoM.Translate(80, 80 + math.Sin(t)*5 + g.jumpOffset) // 应用悬浮和跳跃偏移

	screen.DrawImage(g.img, op)

	// 绘制菜单
	if g.showMenu {
		vector.FillRect(screen, float32(g.menuX), float32(g.menuY), 60, 30, color.RGBA{30, 30, 30, 220}, true)
		vector.StrokeRect(screen, float32(g.menuX), float32(g.menuY), 60, 30, 1, color.RGBA{120, 120, 120, 255}, true)
		vector.FillRect(screen, float32(g.menuX+40), float32(g.menuY+10), 10, 10, color.RGBA{255, 80, 80, 255}, true)
	}
}

func (g *Game) Layout(w, h int) (int, int) { return 160, 160 }

// 去除背景色并转换为 Ebiten 图像
func loadPetImage(path string) (*ebiten.Image, error) {
	f, err := os.Open(path)
	if err != nil { return nil, err }
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil { return nil, err }
	log.Println("成功加载图片，格式为:", format)

	b := img.Bounds()
	res := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			if r > 0xF000 && g > 0xF000 && b > 0xF000 {
				res.Set(x, y, image.Transparent)
			} else {
				res.Set(x, y, c)
			}
		}
	}
	return ebiten.NewImageFromImage(res), nil
}

func main() {
	petImg, err := loadPetImage("pet.png")
	if err != nil { log.Fatal(err) }

	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowSize(160, 160)

	err = ebiten.RunGameWithOptions(&Game{img: petImg}, &ebiten.RunGameOptions{
		ScreenTransparent: true,
		SkipTaskbar:       true,
		InitUnfocused:     true,
	})
	if err != nil { log.Fatal(err) }
}
