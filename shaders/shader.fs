#version 460 compatibility

in vec3 Normal;
in vec3 FragPos;

out vec4 FragColor;

uniform vec3 objectColor;
uniform vec3 lightColor;
uniform vec3 lightPos;
uniform vec3 viewPos;

void main()
{
    // Ambient lighting is an always present constant light, apply a small constant
    // to the light color and put it to the object.
    float ambientStrength = 0.1;
    vec3 ambient = ambientStrength * lightColor;

    // For diffuse lighting, we need to calculate the angle between the light source and the fragment
    // to know how strongly it's affecting the fragment, and how strong to show colors.
    // First, get the direction vector between the light source and fragment by subtracting their position
    // vectors (always normalized in light calculations b/c we only care about the direction of things)
    vec3 norm = normalize(Normal);
    vec3 lightDir = normalize(lightPos - FragPos);
    // The dot product of the normal and light direction give the diffuse impact factor
    float diffuseFactor = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = diffuseFactor * lightColor;

    // Specular lighting is found by calculating the light reflection angle and the closer the angle between
    // that and the viewer (camera), the stronger the specular effect.
    float specularStrength = 0.5;
    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 reflectDir = reflect(-lightDir, norm);
    // calculat the specular component. Shininess effects how scattered the effect is on the object (higher, less scattered)
    int shininess = 32;
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), shininess);
    vec3 specular = specularStrength * spec * lightColor;

    vec3 result = (ambient + diffuse + specular) * objectColor;
    FragColor = vec4(result, 1.0);
}
