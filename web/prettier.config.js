/** @type {import("prettier").Config} */
const config = {
	// defaults
	semi: true,
	singleQuote: false,
	bracketSpacing: true,
	bracketSameLine: false,
	arrowParens: "always",

	// custom options
	// useTabs: true,   // set by .editorconfig
	// printWidth: 100, // set by .editorconfig
	quoteProps: "consistent",
	trailingComma: "none",
	objectWrap: "collapse",

	// plugins
	plugins: ["prettier-plugin-svelte", "prettier-plugin-tailwindcss"],
	overrides: [{ files: "*.svelte", options: { parser: "svelte" } }],
	tailwindStylesheet: "./src/routes/layout.css"
};

export default config;
