package main

import "fmt"
import "sort"

import "time"
import "flag"
import "image"
import "image/color"
import "image/png"
import _ "image/jpeg"
import "image/draw"

import gocolor "./go-color"



import "os"

type Colors []color.Color;

func getLightness(c *color.Color) float64 {
  r, g, b, _ := (*c).RGBA()
  hsl := gocolor.RGB{float64(r) / 65536,float64(g) / 65536,float64(b) / 65536}.ToHSL()

  return hsl.L;

}

// This is a sort interface for some reason, the Len / Swap functions can take
// 'Colors' as the type and don't need ByLightness, but ByLightness would also
// work and make sense
type ByLightness struct {
  Colors
}

func (s Colors) Len() int {
  return len(s);
}

func (s Colors) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s ByLightness) Less(i, j int) bool {
  l1 := getLightness(&s.Colors[i]);
  l2 := getLightness(&s.Colors[j]);
  return l1 > l2;
}

func sort_from(in_img image.Image, out_img *image.RGBA, col, start, end int) {
  if end - start > 100 {
      delta := end - start;
      chunk_size := (delta / 10.0)

      for i := 0; i < delta / chunk_size; i++ { 
        go sort_from(in_img, out_img, col, start + (i * chunk_size), start + ((i + 1) * chunk_size))
      }
      return
  }

  pixels := make([]color.Color, end - start);

  for i := start; i < end; i++ {
    pixels[i - start] = in_img.At(col, i);
  }

  sort.Sort(ByLightness{pixels});

  for i, v := range pixels {
    out_img.Set(col, i + start, v);
  }

}

func save_image(im image.Image) {
  w, _ := os.Create("output.png")

  png.Encode(w, im);

}


func sort_image(im image.Image) {

  width, height := im.Bounds().Max.X, im.Bounds().Max.Y;

  out_img := image.NewRGBA(image.Rect(0, 0, width, height));

  // TODO: flags
  var black_threshold float64 = 0.25;
  var white_threshold float64 = 0.9;

  defer save_image(out_img);

  b := im.Bounds()
  draw.Draw(out_img, out_img.Bounds(), im, b.Min, draw.Src)

  for i := 0; i < width; i++ {
    start := -1;
    end := -1;

    for j := 0; j < height - 1; j++ {
      px := im.At(i, j)
      lightness := getLightness(&px)

      if lightness < black_threshold || lightness > white_threshold {
        if start == -1 {
          start = j;
          continue;
        }

        end = j;

        if end - start > 1 {
          go sort_from(im, out_img, i, start, end);
        }

        start = end;
        end = -1;

      }

    }
  }
}

func main() {
  flag.Parse();
  args := flag.Args()

  file, err := os.Open(args[0]);
  start := time.Now()
  fmt.Println("Opening file", args[0]);
  if err != nil {

    fmt.Println("Couldn't open file", err);
    return;
  }

  img, _, err := image.Decode(file)

  end := time.Now()

  fmt.Println("Opened file, took", end.Sub(start));

  if err != nil {
    fmt.Println("Couldn't open image", err);
    return;
  }

  start = time.Now()
  sort_image(img);
  end = time.Now()
  fmt.Println("Sorted image, took", end.Sub(start));
}
