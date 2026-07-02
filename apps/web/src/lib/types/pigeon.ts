export interface Pigeon {
    id: number;
    name?: string;
    sex?: 'M' | 'F';
    race?: string;
    bandNumber?: string;
    birthDate?: Date;
    captureDate?: Date;
    tags?: string[];
}