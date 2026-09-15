export namespace model {
	
	export class Platform {
	    os: string;
	    architecture: string;
	    variant?: string;
	
	    static createFrom(source: any = {}) {
	        return new Platform(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.architecture = source["architecture"];
	        this.variant = source["variant"];
	    }
	}
	export class JobImage {
	    image: string;
	    targetTag?: string;
	    platform: Platform;

	    static createFrom(source: any = {}) {
	        return new JobImage(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.image = source["image"];
	        this.targetTag = source["targetTag"];
	        this.platform = this.convertValues(source["platform"], Platform);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Job {
	    id: string;
	    name: string;
	    pack?: boolean;
	    outputDir?: string;
	    images: JobImage[];

	    static createFrom(source: any = {}) {
	        return new Job(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.pack = source["pack"];
	        this.outputDir = source["outputDir"];
	        this.images = this.convertValues(source["images"], JobImage);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
			                a[key] = new classs(a[key]);
			            }
			            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DownloadRequest {
	    image: string;
	    targetTag?: string;
	    outputDir: string;
	    platform: Platform;
	    registry: string;
	    username: string;
	    password: string;
	    insecure: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DownloadRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.image = source["image"];
	        this.targetTag = source["targetTag"];
	        this.outputDir = source["outputDir"];
	        this.platform = this.convertValues(source["platform"], Platform);
	        this.registry = source["registry"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.insecure = source["insecure"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DownloadTask {
	    id: string;
	    image: string;
	    outputPath: string;
	    platform: Platform;
	    status: string;
	    // Go type: time
	    startedAt: any;
	    // Go type: time
	    finishedAt?: any;
	    error?: string;
	    digest?: string;
	    size?: number;
	
	    static createFrom(source: any = {}) {
	        return new DownloadTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.image = source["image"];
	        this.outputPath = source["outputPath"];
	        this.platform = this.convertValues(source["platform"], Platform);
	        this.status = source["status"];
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.finishedAt = this.convertValues(source["finishedAt"], null);
	        this.error = source["error"];
	        this.digest = source["digest"];
	        this.size = source["size"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LayerCacheInfo {
	    path: string;
	    files: number;
	    bytes: number;
	    images: number;
	    unreferencedFiles: number;
	    unreferencedBytes: number;

	    static createFrom(source: any = {}) {
	        return new LayerCacheInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.files = source["files"];
	        this.bytes = source["bytes"];
	        this.images = source["images"];
	        this.unreferencedFiles = source["unreferencedFiles"];
	        this.unreferencedBytes = source["unreferencedBytes"];
	    }
	}
	export class CachedLayerImage {
	    image: string;
	    platform: string;
	    digest?: string;

	    static createFrom(source: any = {}) {
	        return new CachedLayerImage(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.image = source["image"];
	        this.platform = source["platform"];
	        this.digest = source["digest"];
	    }
	}
	export class CachedLayer {
	    digest: string;
	    size: number;
	    partial: boolean;
	    referenced: boolean;
	    images: CachedLayerImage[];

	    static createFrom(source: any = {}) {
	        return new CachedLayer(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.digest = source["digest"];
	        this.size = source["size"];
	        this.partial = source["partial"];
	        this.referenced = source["referenced"];
	        this.images = this.convertValues(source["images"], CachedLayerImage);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CachedImage {
	    image: string;
	    digest?: string;
	    platform: string;
	    updatedAt?: string;
	    layerCount: number;
	    cachedCount: number;

	    static createFrom(source: any = {}) {
	        return new CachedImage(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.image = source["image"];
	        this.digest = source["digest"];
	        this.platform = source["platform"];
	        this.updatedAt = source["updatedAt"];
	        this.layerCount = source["layerCount"];
	        this.cachedCount = source["cachedCount"];
	    }
	}
	export class LayerCacheInventory {
	    path: string;
	    files: number;
	    bytes: number;
	    images: number;
	    unreferencedFiles: number;
	    unreferencedBytes: number;
	    layers: CachedLayer[];
	    imageRecords: CachedImage[];

	    static createFrom(source: any = {}) {
	        return new LayerCacheInventory(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.files = source["files"];
	        this.bytes = source["bytes"];
	        this.images = source["images"];
	        this.unreferencedFiles = source["unreferencedFiles"];
	        this.unreferencedBytes = source["unreferencedBytes"];
	        this.layers = this.convertValues(source["layers"], CachedLayer);
	        this.imageRecords = this.convertValues(source["imageRecords"], CachedImage);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileConfig {
	    registry: string;
	    username?: string;
	    password?: string;
	    hasPassword?: boolean;
	    insecure?: boolean;
	    outputDir: string;
	
	    static createFrom(source: any = {}) {
	        return new FileConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.registry = source["registry"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.hasPassword = source["hasPassword"];
	        this.insecure = source["insecure"];
	        this.outputDir = source["outputDir"];
	        this.jobs = this.convertValues(source["jobs"], Job);
	    }

	    jobs?: Job[];

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ImageReference {
	    registry: string;
	    repository: string;
	    tag: string;
	    digest?: string;
	    raw: string;
	
	    static createFrom(source: any = {}) {
	        return new ImageReference(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.registry = source["registry"];
	        this.repository = source["repository"];
	        this.tag = source["tag"];
	        this.digest = source["digest"];
	        this.raw = source["raw"];
	    }
	}
	
	export class RegistryConfig {
	    registry: string;
	    username: string;
	    password: string;
	    insecure: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RegistryConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.registry = source["registry"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.insecure = source["insecure"];
	    }
	}
	export class Result {
	    ok: boolean;
	    message: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}
	export class TestRegistryResult {
	    ok: boolean;
	    registryReachable: boolean;
	    authSuccess: boolean;
	    harborReachable: boolean;
	    message: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TestRegistryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.registryReachable = source["registryReachable"];
	        this.authSuccess = source["authSuccess"];
	        this.harborReachable = source["harborReachable"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}

}

