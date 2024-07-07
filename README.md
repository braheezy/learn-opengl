# Learn OpenGL with Go
[Learn OpenGL](https://learnopengl.com) is an online book to learn OpenGL programming. It uses C because OpenGL is written in C. I did the project in Go because I already know the syntax and the build tooling is much easier.

The [`go-gl`](https://github.com/go-gl) project is used to provide Go bindings to the underlying C OpenGL libraries. If you want to run the code in this repository, follow [this guide](https://github.com/go-gl/glfw?tab=readme-ov-file#installation) to ensure the OpenGL dependencies are installed.

## Repository Organization
`main` follows the main lesson plan.

There are branches like `hello-triangle-ex1` that contain solutions to exercises presented at the end of some chapters. The example give is the Exercise 1 solution for the "Hello Triangle" chapter.

`screenshots` holds snaps of the graphics at various points in the journey. I thought it would be cool to see the visual progression over time. They are named in the order I complete that version of the program. There are commits on `main` or branches that have the code for each picture.
