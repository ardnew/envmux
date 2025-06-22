#!/usr/bin/perl

use strict;
use warnings;

use Cwd 'realpath';
use File::Basename;
use File::Spec;
use File::Temp qw(tempdir);
use IPC::Open2;
use Getopt::Long qw(GetOptions);

sub usage {
  sprintf "Usage: %s [-d|--debug] [-n|--normalize] [-t|--transient] [-c|--testcase <num>]\n", basename($0)
}

# Parse command-line options.
GetOptions(
  \our %opts,
  'debug|d!',
  'normalize|n!',
  'transient|t!',
  'testcase|c=i',
) or die usage();

# Determine relative paths/names from the module and command-line tool.
my $self = realpath($0);
my $path = dirname($self);
my @info = map {
  `go list -f "$_" "$path"`
} (
  '{{.Module.Dir}}',
  '{{.Target}}',
);
chomp @info;

my $root = $info[0];
my $name = basename($info[1]);

# If transient, use a temp file as executable.
my $bin = File::Spec->catfile(
  $opts{'transient'} ? tempdir(CLEANUP => 1) : $path,
  $name,
);
printf "executable: %s\n", $bin;

unless (-d $root) {
  my $fail = $root || dirname($0);
  die "error: module root not found: $fail\n";
}

# Regenerate the lexer from alecthomas/participle/v2.
system("go", "generate", "$root/...") == 0 or do {
  die "error: go generate failed\n";
};

# Build the command-line tool.
system("go", "build", "-v", "-o", $bin) == 0 or do {
  die "error: failed to build $name\n";
};

# -----------------------------------------------------------------------------
# The remaining code is for pretty-printing the results of test cases
# defined at the bottom of this file.
#
#   -- Kudos to copilot for the majority of this tedium.
# -----------------------------------------------------------------------------

sub trim { my $s = shift; $s =~ s/^\s+|\s+$//g; $s }

my %rgb = (
  rst => "\033[0m",
  txt => "\033[4;90m",  # Dark Gray (Underlined)
  num => "\033[3;37m",  # Light Gray (Italic)
  imp => "\033[0;35m",  # Magenta
  box => "\033[0;90m",  # Dark Gray
  err => "\033[0;31m",  # Red
  war => "\033[0;33m",  # Yellow
  i   => "\033[0;36m",  # Cyan
  o   => "\033[0;32m",  # Green
);

sub color {
  my ($key, $str, $rep, $pat, $app) = @_;
  my ($out) = ("", "");

  return unless exists $rgb{$key};

  $out = defined $str
    ? $rgb{$key} . $str . $rgb{rst}  # color with reset.
    : $rgb{$key};  # color without text.

  if (defined $rep and exists $rgb{$rep}) {
    if (defined $pat) {
      $out =~ s/($pat)/$rgb{$rep}${1}$rgb{$key}/g;
    } else {
      $out .= $rgb{$rep};  # Append replacement color.
    }
  }

  if (defined $app and exists $rgb{$app}) {
    $out .= $rgb{$app};  # Append additional color.
  }

  $out
}

my $num = 0;
print for map {
  my ($ref, $fhi, $fho) = ($_, undef, undef);

  die "invalid test case: $ref"
    unless ref($ref) eq 'HASH' and exists $ref->{'def'};

  my ($def, @arg) = ($ref->{'def'}, @{ $ref->{'arg'} || ['default'] });

  my ($dbg) = $opts{'debug'} ? "-vv " : "";

  print $dbg;

  my $pid = open2($fho, $fhi, "${bin} ${dbg}-s - @{arg} 2>&1") or die "$!\n";

  $def = trim($def) if $opts{'normalize'};

  print $fhi $def;
  close $fhi;
  my $res = do { local $/; <$fho> };
  close $fho;
  waitpid($pid, 0);

  my $err = $? >> 8;

  ($def, $res) = ( "\n${def}\n", "\n".trim($res)."\n" )
    if $opts{'normalize'};

  my $sfi = "   │    ".$def;
  $sfi =~ s/\n/\n   │    /g;  # Indent multiline input

  my $sfo = "   │    ".$res;
  $sfo =~ s/\n/\n   │    /g;  # Indent multiline output

  my @hdr = map {
    sprintf("%s%s",
      @$_ > 0
        ? " ".do {
            my ($h, @a) = @$_;
            join " ",
              @a ? (color('txt' => $h), color('imp' => join(', ', @a), 'box'))
                 : (color('txt' => $h, 'box'))
          }." "
        : "",
      "─" x (50 - (@$_ ? $#$_ ? 3 : 2 : 0) - (@$_ ? length($$_[0].join(", ", @$_[1..$#$_])) : 0)),
    )
  } (
    ["Namespace definitions"],
    ["Environment(s)", @arg],
    [],
  );

  my $lbl = color('num' => sprintf("%2d", exists $opts{'testcase'} ? $num : ++$num), 'box');
  my $inp = color('i' => $sfi, 'box' => qr/│/, 'box');
  my $out = color(($err ? 'err' : 'o') => $sfo, 'box' => qr/│/, 'box');

  sprintf <<EOF, $lbl, $hdr[0], $inp, $hdr[1], $out, $hdr[2];
%s ┌─%s
%s
   ├─%s
%s
   └─%s

EOF
} grep {
  ( not exists $opts{'testcase'} ) || ++$num == $opts{'testcase'}
} (

#   ┌───────────┐
#   │ TEST CASES │
#   └───────────┘

  {
    arg => [],
    def => <<___,
default() {foo=1+2;}
___
  },
  {
    arg => [],
    def => <<___,
default() { foo = 1+2
;
 };
___
  },
  {
    arg => [],
    def => <<___,
default() { foo = { 1+2; }; }
___
  },
  {
    arg => [],
    def => <<___,

default() {

  foo =
    1+2
    ;

}
___
  },
  {
    arg => ['foo@bar'],
    def => <<___,
default {
  foo = 1+2;
}
;;;;;;; foo\@bar <default> {
  bar = 3+4;
}
___
  },
)
