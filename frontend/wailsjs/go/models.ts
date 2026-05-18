export namespace main {

	export class BackendStatus {
	    running: boolean;
	    base_url: string;
	    binary_path: string;
	    cli_version: string;
	    token_detected: boolean;
	    spawned_by_app: boolean;
	    message: string;

	    static createFrom(source: any = {}) {
	        return new BackendStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.base_url = source["base_url"];
	        this.binary_path = source["binary_path"];
	        this.cli_version = source["cli_version"];
	        this.token_detected = source["token_detected"];
	        this.spawned_by_app = source["spawned_by_app"];
	        this.message = source["message"];
	    }
	}
	export class BulkActionResult {
	    attempted: number;
	    succeeded: number;
	    errors: string[];

	    static createFrom(source: any = {}) {
	        return new BulkActionResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.attempted = source["attempted"];
	        this.succeeded = source["succeeded"];
	        this.errors = source["errors"];
	    }
	}
	export class DownloadItem {
	    id: string;
	    url: string;
	    filename: string;
	    dest_path: string;
	    total_size: number;
	    downloaded: number;
	    progress: number;
	    speed: number;
	    status: string;
	    eta: number;
	    connections: number;
	    error?: string;
	    time_taken: number;
	    avg_speed: number;

	    static createFrom(source: any = {}) {
	        return new DownloadItem(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.filename = source["filename"];
	        this.dest_path = source["dest_path"];
	        this.total_size = source["total_size"];
	        this.downloaded = source["downloaded"];
	        this.progress = source["progress"];
	        this.speed = source["speed"];
	        this.status = source["status"];
	        this.eta = source["eta"];
	        this.connections = source["connections"];
	        this.error = source["error"];
	        this.time_taken = source["time_taken"];
	        this.avg_speed = source["avg_speed"];
	    }
	}

}

