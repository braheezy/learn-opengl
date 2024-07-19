# Screenshots
At the end of each chapter and for each relevant exercise within, I took a capture of what I was drawing. I wrote [`capture.sh`](../capture.sh) to automate this.

Enjoy the journey.
---
![Hello Window](./0_hello-window.png)

In the beginning, there was a window. The dimensions are backwards but I press on.

---
![Hello Triangle](./1_hello-triangle.png)
***Hello Triangle*** is the ceremonious first-program that a graphics programmer must perform. The absolute basics are explained to put 4 vertices together and draw two triangles. My gaming GPU is sadden by how useless it feels.

---
![Hello Triangle Exercise 1](./2_hello-triangle-ex1.png)
> **Hello Triangle Exercise 1**: Try to draw 2 triangles next to each other using glDrawArrays by adding more vertices to your data

And so 2 triangles are drawn. Wireframe mode is gone.

---
![Hello Triangle Exercise 3](./3_hello-triangle-ex3.png)
> **Hello Triangle Exercise 3**: Create two shader programs where the second program uses a different fragment shader that outputs the color yellow; draw both triangles again where one outputs the color yellow

Duplicating all the existing code and changing the color bits.

---
![Shaders](./4_shaders.png)

Finally, I know what shaders are! Tiny(ish) programs that live on the GPU and, in the case of graphics programming, help process the graphics pipeline. Here, the pixels at the tips of triangle are set to RGB, respectively, by the fragment shader and the rest of the colors are automatically interpolated based on position.

---
![Shaders Exercise 1](./5_shaders-ex1.png)

> **Shaders Exercise 1**: Adjust the vertex shader so that the triangle is upside down.

This LearnOpenGL site is good at making small exercises to reinforce the learning.

---
![Shaders Exercise 2](./6_shaders-ex2.png)

> **Shaders Exercise 2**: Specify a horizontal offset via a uniform and move the triangle to the right side of the screen in the vertex shader using this offset value

Uniforms are like global variables across all shader programs and they can be set from outside the shader program.

---
![Shaders Exercise 3](./7_shaders-ex3.png)

> **Shaders Exercise 3**: Output the vertex position to the fragment shader using the out keyword and set the fragment's color equal to this vertex position (see how even the vertex position values are interpolated across the triangle). Once you managed to do this; try to answer the following question: why is the bottom-left side of our triangle black?

The black pixels happen because color values must be between 0.0 and 1.0. The blacks pixels all lie in the negative part of the XY plane and are clamped to 0.0, or black.

---
![Camera](./18_camera.mp4)

With extensive matrix maths, a fly camera is born. In this video, I have to toggle the cursor capture behavior so I can run the app then record it with OBS. WASD and the mouse are freely navigating and zooming the scene.
