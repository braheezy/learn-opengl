#version 460 compatibility
layout (location = 0) in vec3 aPos;   // the position variable has attribute position 0
layout (location = 1) in vec3 aNormal;

out vec3 Normal;
out vec3 FragPos;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;
uniform vec3 viewPos;

void main()
{
    // Get the fragment's position in world space
    FragPos = vec3(model * vec4(aPos, 1.0));
    Normal =  mat3(transpose(inverse(model))) * aNormal;
    // note that we read the multiplication from right to left
    gl_Position = projection * view * model * vec4(FragPos, 1.0);
}
