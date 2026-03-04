export namespace main {
	
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

