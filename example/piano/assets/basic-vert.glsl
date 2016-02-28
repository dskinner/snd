#version 100

attribute vec4 position;
uniform mat4 world;
uniform mat4 view;
uniform mat4 proj;

void main() {
  gl_Position = position * world * view * proj;
}
