/*
Copyright 2019 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package executor

import (
	"strings"

	apis "github.com/aamir-tiwari-sumo/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/aamir-tiwari-sumo/maya/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/klog"

	upgrader "github.com/aamir-tiwari-sumo/maya/pkg/upgrade/upgrader"
	errors "github.com/pkg/errors"
)

// JivaVolumeOptions stores information required for jiva volume upgrade
type JivaVolumeOptions struct {
	pvName string
}

var (
	jivaVolumeUpgradeCmdHelpText = `
This command upgrades one or many Jiva Persistent Volume
`
	jivaVolumeUpgradeCmdExampleText = `  # Upgrade one volume at a time
  upgrade jiva-volume --pv-name <pv-name> --options...

  # Upgrade multiple volumes at a time
  upgrade jiva-volume <pv-name>... --options...`
)

// NewUpgradeJivaVolumeJob upgrade a Jiva Volume
func NewUpgradeJivaVolumeJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jiva-volume",
		Short:   "Upgrade Jiva Volume",
		Long:    jivaVolumeUpgradeCmdHelpText,
		Example: jivaVolumeUpgradeCmdExampleText,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.RunJivaVolumeUpgradeChecks(args), util.Fatal)
			options.resourceKind = "jivaVolume"
			if options.jivaVolume.pvName != "" {
				singleJivaUpgrade(cmd)
			}
			if len(args) != 0 {
				bulkJivaUpgrade(cmd, args)
			}
		},
	}

	cmd.Flags().StringVarP(&options.jivaVolume.pvName,
		"pv-name", "",
		options.jivaVolume.pvName,
		"jiva persistent volume name to be upgraded, as obtained using: kubectl get pv")

	return cmd
}

func singleJivaUpgrade(cmd *cobra.Command) {
	util.CheckErr(options.RunPreFlightChecks(cmd), util.Fatal)
	util.CheckErr(options.InitializeDefaults(cmd), util.Fatal)
	util.CheckErr(options.RunJivaVolumeUpgrade(cmd), util.Fatal)
}

func bulkJivaUpgrade(cmd *cobra.Command, args []string) {
	for _, name := range args {
		options.jivaVolume.pvName = name
		singleJivaUpgrade(cmd)
	}
}

// RunJivaVolumeUpgradeChecks will ensure the sanity of the jiva upgrade options
func (u *UpgradeOptions) RunJivaVolumeUpgradeChecks(args []string) error {
	if len(strings.TrimSpace(u.jivaVolume.pvName)) == 0 && len(args) == 0 {
		return errors.Errorf("Cannot execute upgrade job:" +
			" neither pv-name flag is set nor pv name list is provided")
	}

	return nil
}

// RunJivaVolumeUpgrade upgrades the given Jiva Volume.
func (u *UpgradeOptions) RunJivaVolumeUpgrade(cmd *cobra.Command) error {

	klog.V(4).Infof("Started upgrading %s{%s} from %s to %s",
		u.resourceKind,
		u.jivaVolume.pvName,
		u.fromVersion,
		u.toVersion)

	if apis.IsCurrentVersionValid(u.fromVersion) && apis.IsDesiredVersionValid(u.toVersion) {
		klog.Infof("Upgrading to %s", u.toVersion)
		err := upgrader.Exec(u.fromVersion, u.toVersion,
			u.resourceKind,
			u.jivaVolume.pvName,
			u.openebsNamespace,
			u.imageURLPrefix,
			u.toVersionImageTag)
		if err != nil {
			klog.Error(err)
			return errors.Wrapf(err, "Failed to upgrade %s{%s}",
				u.resourceKind,
				u.jivaVolume.pvName)
		}
	} else {
		return errors.Errorf("Invalid from version %s or to version %s", u.fromVersion, u.toVersion)
	}
	klog.V(4).Infof("Successfully upgraded %s{%s} from %s to %s",
		u.resourceKind,
		u.jivaVolume.pvName,
		u.fromVersion,
		u.toVersion)
	return nil
}
