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
    vec3 direction;
    float cutoff;
    float outerCutoff;

    vec3 ambient;
    vec3 diffuse;
    vec3 specular;

    // attentuation
    float constant;
    float linear;
    float quadratic;
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
    vec3 ambient = light.ambient * texture(material.diffuse, TexCoords).rgb;

    // For diffuse lighting, we need to calculate the angle between the light source and the fragment
    // to know how strongly it's affecting the fragment, and how strong to show colors.
    // First, get the direction vector between the light source and fragment by subtracting their position
    // vectors (always normalized in light calculations b/c we only care about the direction of things)
    vec3 norm = normalize(Normal);
    vec3 lightDir = normalize(light.position - FragPos);
    float diff = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = light.diffuse * diff * texture(material.diffuse, TexCoords).rgb;

    // specular
    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 reflectDir = reflect(-lightDir, norm);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), material.shininess);
    vec3 specular = light.specular * spec * texture(material.specular, TexCoords).rgb;

    // spotlight (soft edges)
    float theta = dot(lightDir, normalize(-light.direction));
    float epsilon = (light.cutoff - light.outerCutoff);
    float intensity = clamp((theta - light.outerCutoff) / epsilon, 0.0, 1.0);
    diffuse  *= intensity;
    specular *= intensity;

    // attenuation
    float distance    = length(light.position - FragPos);
    float attenuation = 1.0 / (light.constant + light.linear * distance + light.quadratic * (distance * distance));
    ambient  *= attenuation;
    diffuse   *= attenuation;
    specular *= attenuation;

    vec3 result = ambient + diffuse + specular;
    FragColor = vec4(result, 1.0);
}
