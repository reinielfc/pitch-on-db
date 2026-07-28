export class ApiError extends Error {
	constructor(
		public status: number,
		public body: unknown,
		message?: string
	) {
		super(message ?? "API request failed");
		this.name = "ApiError";
	}
}
