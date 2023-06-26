{ lib, pkgs, ... }:
let
  lockBackground = builtins.path {
    path = ./lock-background.jpg;
    name = "slick-lock-background";
  };
in
{
  imports = [
    ./hardware-configuration.nix
  ];

  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;
  boot.supportedFilesystems = [ "zfs" "ntfs" ];

  boot.kernel.sysctl."net.ipv4.conf.all.arp_ignore" = 1;
  boot.kernel.sysctl."net.ipv4.conf.all.arp_announce" = 1;

  networking = {
    firewall.enable = false;
    nameservers = [ "1.1.1.1" "8.8.8.8" ];
    hostName = "noah-nixos-desktop";
    hostId = "8425e349";
    networkmanager = {
      enable = true;
      dns = "none";
      extraConfig = ''
        [connection-wifi-wlp6s0]
        match-device=interface-name:wlp6s0
        ipv4.route-metric=10

        [connection-eth-eno1]
        match-device=interface-name:eno1
        ipv4.route-metric=100
      '';
    };
    extraHosts = ''
      127.0.0.1 sourcegraph.test
    '';
  };

  time.timeZone = "Europe/Dublin";

  i18n.defaultLocale = "en_US.UTF-8";
  #console = {
  #  font = "Lat2-Terminus16";
  #  keyMap = "us";
  #  useXkbConfig = true; # use xkbOptions in tty.
  #};

  fonts = {
    fonts = with pkgs; [
      noto-fonts
      noto-fonts-cjk
      noto-fonts-emoji
      noto-fonts-emoji-blob-bin
      joypixels
      dejavu_fonts
      fira-code
      font-awesome
      font-awesome_5
      monostroom
      code2000
    ];
    fontDir.enable = true;
    fontconfig = {
      enable = true;
      hinting = {
        enable = true;
        autohint = false;
        style = "hintslight";
      };
      subpixel = {
        rgba = "rgb";
        lcdfilter = "default";
      };
      defaultFonts = {inherit (import ./fonts.nix) monospace sansSerif emoji serif; };
    };
  };

  services.zfs.autoSnapshot.enable = true;
  services.zfs.autoScrub.enable = true;

  location = {
    latitude = 52.3379105;
    longitude = -6.4560767;
  };

  services.geoclue2.enable = true;
  services.redshift = {
    enable = true;
    temperature.night = 4500;
  };
  services.xserver = {
    enable = true;
    exportConfiguration = true;

    windowManager.i3 = {
      enable = true;
      extraPackages = with pkgs; [
        dmenu
        rofi
        polybar
      ];
      extraSessionCommands = ''
        eval $(${pkgs.gnome3.gnome-keyring}/bin/gnome-keyring-daemon --daemonize --components=ssh,secrets)
        export SSH_AUTH_SOCK
      '';
    };

    displayManager.defaultSession = "none+i3";
    displayManager.lightdm = {
      enable = true;
      greeter.enable = true;
      greeters.slick = {
        enable = true;
        draw-user-backgrounds = false;
        extraConfig = ''
          [Greeter]
          background=${lockBackground}
          stretch-background-across-monitors=false
        '';
      };
    };

    dpi = 96;
    videoDrivers = [ "nvidia" ];
    screenSection = ''
      Option "metamodes" "DP-4.8: 2560x1440_60 +0+0 {ForceCompositionPipeline=On, ForceFullCompositionPipeline=On}, HDMI-0: 1920x1080_60 +2560+360"
    '';
  };
  services.picom.enable = true;
  hardware.opengl.enable = true;
  # hardware.nvidia.forceFullCompositionPipeline = true;

  # Configure keymap in X11
  # services.xserver.layout = "us";
  # services.xserver.xkbOptions = {
  #   "eurosign:e";
  #   "caps:escape" # map caps to escape.
  # };

  services.printing = {
    enable = true;
    drivers = with pkgs; [ epson-escpr ];
  };

  sound.enable = true;
  hardware.pulseaudio.enable = true;

  users.mutableUsers = false;
  users.users.noah = {
    isNormalUser = true;
    extraGroups = [ "wheel" "networkmanager" "docker" "audio" ];
    shell = pkgs.fish;
    home = "/home/noah";
    createHome = true;
    hashedPassword = "$y$jFT$4BKFwYX3OJFl9W6Md0cw./$fS16Nf1gFV3PecFbe5LfzCulv4OoLJFKz8nEfXi.pz0";
  };

  environment.sessionVariables = with pkgs; {
    JAVA_8_HOME = "${jdk8}";
    JAVA_11_HOME = "${jdk11}";
    JAVA_17_HOME = "${jdk17}";
    JAVA_19_HOME = "${jdk19}";
    _JAVA_OPTIONS = "-Dawt.useSystemAAFontSettings=lcd";
    JAVA_HOME = "${jdk11}";
    RUST_SRC_PATH = "${pkgs.rust.packages.stable.rustPlatform.rustLibSrc}";
  };

  environment.pathsToLink = [ "/libexec" "/share/nix-direnv" ];
  environment.systemPackages = with pkgs; [
    vim
    git
    git-crypt
    wget
    coreutils
    man
    tree
    mkpasswd
    firefox
    thunderbird
    vscode
    gparted
    pass
    browserpass
    openssh
    gnupg
    pinentry
    pinentry-gnome
    discord
    slack
    jetbrains.idea-ultimate
    jetbrains.idea-community
    jdk8
    jdk11
    jdk17
    jdk19
    kotlin
    go_1_19
    gopls
    python311
    # bazel
    nodejs-19_x
    maven
    gradle
    synergy
    kitty
    nil
    nix-index
    cinnamon.nemo
    cinnamon.mint-themes
    (cinnamon.mint-y-icons.overrideAttrs (oldAttrs: {
      pname = "mint-l-icons";
      src = pkgs.fetchFromGitHub {
        owner = "linuxmint";
        repo = "mint-l-icons";
        rev = "e9fd3cf2d3f3a22647e9a83da9b16538795fddbb";
        sha256 = "sha256-RDozoknjXqzjQxLgOAaD/BH7hhi5mNlW+Vne93aEt0I=";
      };
    }))
    htop
    coursier
    (adapta-gtk-theme.overrideAttrs (oldAttrs: {
      version = "3.95.0.11-custom";
      src = pkgs.fetchFromGitHub {
        owner = "Strum355";
        repo = "adapta-gtk-theme";
        rev = "26dcba1068bd2ce30328df44a911d14471dac030";
        sha256 = "sha256-MHkR3sUNjeO9A751y3jCdyB4OJhBUZA1X/Sxtx8HOcM=";
      };
    }))
    lxappearance
    flameshot
    gnome.zenity
    gnome.gnome-keyring
    gnome.file-roller
    gnome.eog
    playerctl
    fzf
    fishPlugins.foreign-env
    fishPlugins.fzf-fish
    fishPlugins.bobthefish
    wdiff
    diff-so-fancy
    fontconfig
    font-manager
    vlc
    nitrogen
    xclip
    unzip
    file
    gcc
    clang
    telegram-desktop
    zoom-us
    patchelf
    direnv
    nix-direnv
    docker-compose
    graphviz
    jq
  ];

  security.pam.services.lightdm.enableGnomeKeyring = true;
  services.gnome.gnome-keyring.enable = true;

  programs = {
    gnupg.agent = {
      enable = true;
    };
    # nix-ld.enable = true;
    # ssh.startAgent = true;
    fish.enable = true;
    browserpass.enable = true;
    dconf.enable = true;
  };

  #services.passSecretService.enable = true;
  # services.openssh.enable = true;

  virtualisation.docker = {
    rootless.enable = true;
    enable = true;
  };

  nixpkgs.config.joypixels.acceptLicense = true;
  nixpkgs.config.allowUnfree = true;
  nixpkgs.overlays = [
    (self: super: { nix-direnv = super.nix-direnv.override { enableFlakes = true; }; })
  ];
  nix = {
    settings = {
      experimental-features = [ "nix-command" "flakes" ];
      auto-optimise-store = true;
      keep-outputs = true;
    };
    gc.automatic = true;
    optimise.automatic = true;
  };

  system.stateVersion = "22.11";
  system.activationScripts.ldso = lib.stringAfter [ "usrbinenv" ] ''
    mkdir -m 0755 -p /lib64
    ln -sfn ${pkgs.glibc.out}/lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2.tmp
    mv -f /lib64/ld-linux-x86-64.so.2.tmp /lib64/ld-linux-x86-64.so.2 # atomically replace
  '';
}
