pub contract Profile {
  pub event UserNameChanged(address: Address?, name: String)
  pub event UserAvatarChanged(address: Address?, avatar: String)
  pub event UserCoverChanged(address: Address?, cover: String)

  pub let publicPath: PublicPath
  pub let privatePath: StoragePath

  pub resource interface Public {
    pub fun getID(): String
    pub fun getName(): String
    pub fun getAvatar(): String
    pub fun getCover(): String
    pub fun asReadOnly(): Profile.ReadOnly
  }

  pub resource interface Owner {
    pub fun getID(): String
    pub fun getName(): String
    pub fun getAvatar(): String
    pub fun getCover(): String

    pub fun setID(_ id: String) {
      pre {
        id.length <= 15: "ID must be under 15 characters long."
      }
    }
    pub fun setName(_ name: String) {
      pre {
        name.length <= 15: "Names must be under 15 characters long."
      }
    }
    pub fun setAvatar(_ src: String)
    pub fun setCover(_ src: String)
  }

  pub resource Base: Owner, Public {
    access(self) var id: String
    access(self) var name: String
    access(self) var avatar: String
    access(self) var cover: String

    init() {
      self.id = "anon"
      self.name = "Anon"
      self.avatar = ""
      self.cover = ""
    }

    pub fun getID(): String { return self.id }
    pub fun getName(): String { return self.name }
    pub fun getAvatar(): String { return self.avatar }
    pub fun getCover(): String { return self.cover }

    pub fun setID(_ id: String) { self.id = id }

    pub fun setName(_ name: String) {
      self.name = name
      emit UserNameChanged(address: self.owner?.address, name: name)
    }
    pub fun setAvatar(_ src: String) {
      self.avatar = src
      emit UserAvatarChanged(address: self.owner?.address, avatar: src)
    }
    pub fun setCover(_ src: String) {
      self.cover = src
      emit UserCoverChanged(address: self.owner?.address, cover: src)
    }

    pub fun asReadOnly(): Profile.ReadOnly {
      return Profile.ReadOnly(
        address: self.owner?.address,
        id: self.getID(),
        name: self.getName(),
        avatar: self.getAvatar(),
        cover: self.getCover()
      )
    }
  }

  pub struct ReadOnly {
    pub let address: Address?
    pub let id: String
    pub let name: String
    pub let avatar: String
    pub let cover: String

    init(address: Address?, id: String, name: String, avatar: String, cover: String) {
      self.address = address
      self.id = id
      self.name = name
      self.avatar = avatar
      self.cover = cover
    }
  }

  pub fun new(): @Profile.Base {
    return <- create Base()
  }

  pub fun check(_ address: Address): Bool {
    return getAccount(address)
      .getCapability<&{Profile.Public}>(Profile.publicPath)
      .check()
  }

  pub fun fetch(_ address: Address): &{Profile.Public} {
    return getAccount(address)
      .getCapability<&{Profile.Public}>(Profile.publicPath)
      .borrow()!
  }

  pub fun read(_ address: Address): Profile.ReadOnly? {
    if let profile = getAccount(address).getCapability<&{Profile.Public}>(Profile.publicPath).borrow() {
      return profile.asReadOnly()
    } else {
      return nil
    }
  }

  pub fun readMultiple(_ addresses: [Address]): {Address: Profile.ReadOnly} {
    let profiles: {Address: Profile.ReadOnly} = {}
    for address in addresses {
      let profile = Profile.read(address)
      if profile != nil {
        profiles[address] = profile!
      }
    }
    return profiles
  }


  init() {
    self.publicPath = /public/profile
    self.privatePath = /storage/profile

    // To avoid error when we are updating Profile contract
    let oldMinter <- self.account.load<@Base{Public}>(from: self.privatePath)
    destroy oldMinter

    self.account.save(<- self.new(), to: self.privatePath)
    self.account.link<&Base{Public}>(self.publicPath, target: self.privatePath)

    self.account
      .borrow<&Base{Owner}>(from: self.privatePath)!
      .setID("anonId")

    self.account
      .borrow<&Base{Owner}>(from: self.privatePath)!
      .setName("Anon")

    self.account
      .borrow<&Base{Owner}>(from: self.privatePath)!
      .setAvatar("https://avatars.githubusercontent.com/u/80470819?v=4")

    self.account
      .borrow<&Base{Owner}>(from: self.privatePath)!
      .setCover("https://nft-image-bucket.s3.us-west-2.amazonaws.com/e24c3d3644e84e7f0c99035e38cc97e0")
  }
}
