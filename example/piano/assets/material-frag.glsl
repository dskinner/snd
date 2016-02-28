#version 100
#define twopi 6.28318530718
#define sqrt2 1.41421356237
#define onesqrt2 0.70710678118
precision mediump float;

uniform sampler2D tex0;
varying vec2 vtc0;

uniform vec2 icon;
uniform vec4 color;
varying vec3 vpos;

uniform int circle;

// const float iconstep = 0.0234375;
const float offset = 0.00024414062;

void main() {
	if (icon.x >= 0.0) {
		gl_FragColor = texture2D(tex0, vtc0+icon);
		gl_FragColor.rgb = color.rgb;
		gl_FragColor.a *= 0.54; // https://www.google.com/design/spec/style/color.html#color-ui-color-application
	} else if (circle == 1) {
		float dist = length(vpos.xy-0.5);
		if (0.49 <= dist && dist <= 0.5) {
			gl_FragColor = vec4(color.rgb, 0.75);
		} else if (dist < 0.49) {
			gl_FragColor = color;
		} else {
			gl_FragColor = vec4(0.0);
		}
	} else {
		gl_FragColor = color;
	}

	// bottom of material
	// if (vpos.z < 0.0) { 
	// gl_FragColor.rgb *= 0.7;
	// }
}
