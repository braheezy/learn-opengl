#version 460 compatibility
// Collect different color and settings together to define a Material
struct Material {
    // A diffuse map contains info on how each fragment diffuse behavior
    // should be (ambient indirectly too)
    sampler2D diffuse;
    sampler2D specular;
    float shininess;
};
// Define how intense light should be
struct Light {
    vec3 position;

    vec3 ambient;
    vec3 diffuse;
    vec3 specular;
};

uniform Light light;
uniform Material material;
uniform vec3 viewPos;

in vec3 Normal;
in vec3 FragPos;
in vec2 TexCoords;

out vec4 FragColor;

void main()
{
    // Ambient lighting is an always present constant light, apply a small constant
    // to the light color and put it to the object.
    vec3 ambient = light.ambient * vec3(texture(material.diffuse, TexCoords).rgb);

    // For diffuse lighting, we need to calculate the angle between the light source and the fragment
    // to know how strongly it's affecting the fragment, and how strong to show colors.
    // First, get the direction vector between the light source and fragment by subtracting their position
    // vectors (always normalized in light calculations b/c we only care about the direction of things)
    vec3 norm = normalize(Normal);
    vec3 lightDir = normalize(light.position - FragPos);
    // The dot product of the normal and light direction give the diffuse impact factor
    float diffuseFactor = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = light.diffuse * diffuseFactor * vec3(texture(material.diffuse, TexCoords).rgb);

    // Specular lighting is found by calculating the light reflection angle and the closer the angle between
    // that and the viewer (camera), the stronger the specular effect.
    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 reflectDir = reflect(-lightDir, norm);
    // calculat the specular component. Shininess effects how scattered the effect is on the object (higher, less scattered)
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), material.shininess);
    vec3 specular = light.specular * spec * vec3(texture(material.specular, TexCoords).rgb);

    FragColor = vec4(ambient + diffuse + specular, 1.0);
}
