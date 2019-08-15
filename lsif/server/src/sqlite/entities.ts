import { PrimaryGeneratedColumn, Column, Entity, PrimaryColumn, Index, JoinColumn, ManyToOne } from 'typeorm'

//
// Single-repo Entities

@Entity({ name: 'meta' })
export class Meta {
    @PrimaryGeneratedColumn() id!: number
    @Column() lsifVersion!: string
    @Column() sourcegraphVersion!: string
}

@Entity({ name: 'blobs' })
export class Blob {
    @PrimaryColumn() hash!: string
    @Column() value!: string
}

@Entity({ name: 'documents' })
export class Document {
    @PrimaryColumn() hash!: string
    @Column() uri!: string
}

@Entity({ name: 'defs' })
@Index(['scheme', 'identifier'])
export class Def {
    @PrimaryColumn() id!: number
    @Column() scheme!: string
    @Column() identifier!: string
    @Column() startLine!: number
    @Column() endLine!: number
    @Column() startCharacter!: number
    @Column() endCharacter!: number
    @Column() documentHash!: string
    @ManyToOne(type => Document) @JoinColumn() document!: Document
}

@Entity({ name: 'refs' })
@Index(['scheme', 'identifier'])
export class Ref {
    @PrimaryColumn() id!: number
    @Column() scheme!: string
    @Column() identifier!: string
    @Column() startLine!: number
    @Column() endLine!: number
    @Column() startCharacter!: number
    @Column() endCharacter!: number
    @Column() documentHash!: string
    @ManyToOne(type => Document) @JoinColumn() document!: Document
}

@Entity({ name: 'hovers' })
@Index(['scheme', 'identifier'])
export class Hover {
    @PrimaryColumn() id!: number
    @Column() scheme!: string
    @Column() identifier!: string
    @Column() blobHash!: string
    @ManyToOne(type => Blob) @JoinColumn() blob!: Blob
}

//
// Xrepo Entities

@Entity({ name: 'packages' })
@Index(['scheme', 'name', 'version'])
export class Package {
    @PrimaryGeneratedColumn() id!: number
    @Column() scheme!: string
    @Column() name!: string
    @Column() version!: string
    @Column() repository!: string
    @Column() commit!: string
}

@Entity({ name: 'references' })
@Index(['scheme', 'name', 'version'])
export class Reference {
    @PrimaryGeneratedColumn() id!: number
    @Column() scheme!: string
    @Column() name!: string
    @Column() version!: string
    @Column() repository!: string
    @Column() commit!: string
    @Column() filter!: string
}
