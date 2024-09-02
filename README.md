# Learn OpenGL with Go
[Learn OpenGL](https://learnopengl.com) is an online book to learn OpenGL programming. This repository holds my code for all the lessons. The tutorial uses C because OpenGL is written in C. I did the project in Go because I already know the syntax and the build tooling is much easier.

The [`go-gl`](https://github.com/go-gl) project is used to provide Go bindings to the underlying C OpenGL libraries. If you want to run the code in this repository, follow [this guide](https://github.com/go-gl/glfw?tab=readme-ov-file#installation) to ensure the OpenGL dependencies are installed.

## Repository Organization
`main` follows the main lesson plan. The root of the project follows the main lesson so running `go run .` will launch the last lesson of the tutorial, "Text Rendering".

The final project is the game Breakout from scratch. That can be run on `main` by doing `go run ./breakout`.

There are branches like `hello-triangle-ex1` that contain solutions to exercises presented at the end of some chapters. `hello-triangle-ex1` is the Exercise 1 solution for the "Hello Triangle" chapter.

[`screenshots`](./screenshots/README.md) holds snaps of the graphics at various points in the lesson.
